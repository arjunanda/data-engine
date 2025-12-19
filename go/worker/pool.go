package worker

import (
	"context"
	"fmt"
	"sync"
)

// Batch represents a batch of rows to be processed
type Batch struct {
	Rows [][]interface{}
	Err  error
}

// Pool manages a pool of worker goroutines for concurrent processing
type Pool struct {
	workers   int
	batchSize int
	inputCh   chan []interface{}
	batchCh   chan Batch
	errorCh   chan error
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewPool creates a new worker pool
func NewPool(ctx context.Context, workers, batchSize int) *Pool {
	ctx, cancel := context.WithCancel(ctx)
	
	return &Pool{
		workers:   workers,
		batchSize: batchSize,
		inputCh:   make(chan []interface{}, workers*2), // Buffered for backpressure
		batchCh:   make(chan Batch, workers),
		errorCh:   make(chan error, 1),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start starts the worker pool
func (p *Pool) Start(processBatch func(context.Context, [][]interface{}) error) {
	// Start batch accumulator
	p.wg.Add(1)
	go p.accumulator()

	// Start workers
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i, processBatch)
	}
}

// accumulator accumulates rows into batches
func (p *Pool) accumulator() {
	defer p.wg.Done()
	defer close(p.batchCh)

	batch := make([][]interface{}, 0, p.batchSize)

	flush := func() {
		if len(batch) > 0 {
			// Make a copy to avoid race conditions
			batchCopy := make([][]interface{}, len(batch))
			copy(batchCopy, batch)
			
			select {
			case p.batchCh <- Batch{Rows: batchCopy}:
			case <-p.ctx.Done():
				return
			}
			
			batch = batch[:0] // Reset batch
		}
	}

	for {
		select {
		case row, ok := <-p.inputCh:
			if !ok {
				// Input channel closed, flush remaining batch
				flush()
				return
			}
			
			batch = append(batch, row)
			if len(batch) >= p.batchSize {
				flush()
			}

		case <-p.ctx.Done():
			return
		}
	}
}

// worker processes batches
func (p *Pool) worker(id int, processBatch func(context.Context, [][]interface{}) error) {
	defer p.wg.Done()

	for {
		select {
		case batch, ok := <-p.batchCh:
			if !ok {
				return
			}

			if err := processBatch(p.ctx, batch.Rows); err != nil {
				// Send error and cancel context
				select {
				case p.errorCh <- fmt.Errorf("worker %d: %w", id, err):
				default:
				}
				p.cancel()
				return
			}

		case <-p.ctx.Done():
			return
		}
	}
}

// Submit submits a row to the pool
func (p *Pool) Submit(row []interface{}) error {
	select {
	case p.inputCh <- row:
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	}
}

// Close closes the pool and waits for all workers to finish
func (p *Pool) Close() error {
	close(p.inputCh) // Signal no more input
	p.wg.Wait()      // Wait for all workers to finish

	// Check for errors
	select {
	case err := <-p.errorCh:
		return err
	default:
		return nil
	}
}

// Cancel cancels the pool context
func (p *Pool) Cancel() {
	p.cancel()
}
