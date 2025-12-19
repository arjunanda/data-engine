package importer

import (
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

const (
	// MaxXLSXSize is the maximum allowed XLSX file size (100MB)
	MaxXLSXSize = 100 * 1024 * 1024
)

// XLSXImporter handles XLSX file imports with streaming
type XLSXImporter struct {
	filePath string
	file     *excelize.File
	rows     *excelize.Rows
	columns  []string
}

// NewXLSXImporter creates a new XLSX importer
func NewXLSXImporter(filePath string) *XLSXImporter {
	return &XLSXImporter{
		filePath: filePath,
	}
}

// Open opens the XLSX file and validates size limits
func (x *XLSXImporter) Open() ([]string, error) {
	// Check file size
	info, err := os.Stat(x.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if info.Size() > MaxXLSXSize {
		return nil, fmt.Errorf("XLSX file too large: %d bytes (max %d bytes). Please convert to CSV for large files", 
			info.Size(), MaxXLSXSize)
	}

	// Open XLSX file
	file, err := excelize.OpenFile(x.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open XLSX file: %w", err)
	}
	x.file = file

	// Get first sheet name
	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		file.Close()
		return nil, fmt.Errorf("no sheets found in XLSX file")
	}

	sheetName := sheets[0]
	fmt.Fprintf(os.Stderr, "[INFO] Reading sheet: %s\n", sheetName)

	// Open streaming reader
	rows, err := file.Rows(sheetName)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to create row iterator: %w", err)
	}
	x.rows = rows

	// Read header row
	if !rows.Next() {
		rows.Close()
		file.Close()
		return nil, fmt.Errorf("empty sheet")
	}

	header, err := rows.Columns()
	if err != nil {
		rows.Close()
		file.Close()
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	x.columns = header
	return header, nil
}

// NextRow reads the next row from the XLSX file
func (x *XLSXImporter) NextRow() ([]interface{}, error) {
	if !x.rows.Next() {
		if err := x.rows.Error(); err != nil {
			return nil, fmt.Errorf("row iteration error: %w", err)
		}
		return nil, fmt.Errorf("EOF")
	}

	cols, err := x.rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to read row: %w", err)
	}

	// Convert to interface slice and pad if necessary
	row := make([]interface{}, len(x.columns))
	for i := 0; i < len(x.columns); i++ {
		if i < len(cols) {
			row[i] = cols[i]
		} else {
			row[i] = nil
		}
	}

	return row, nil
}

// Close closes the XLSX file
func (x *XLSXImporter) Close() error {
	if x.rows != nil {
		x.rows.Close()
	}
	if x.file != nil {
		return x.file.Close()
	}
	return nil
}
