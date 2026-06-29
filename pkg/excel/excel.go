package excel

import (
	"io"

	"github.com/xuri/excelize/v2"
)

type Exporter struct {
	file *excelize.File
}

func NewExporter() *Exporter {
	return &Exporter{
		file: excelize.NewFile(),
	}
}

func (e *Exporter) SetSheet(name string) {
	e.file.NewSheet(name)
}

func (e *Exporter) SetHeaders(sheet string, headers []string) {
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		e.file.SetCellValue(sheet, cell, h)
	}
}

func (e *Exporter) SetRow(sheet string, row int, values []interface{}) {
	for i, v := range values {
		cell, _ := excelize.CoordinatesToCellName(i+1, row)
		e.file.SetCellValue(sheet, cell, v)
	}
}

func (e *Exporter) WriteToWriter(w io.Writer) error {
	return e.file.Write(w)
}

func (e *Exporter) Save(path string) error {
	return e.file.SaveAs(path)
}

func (e *Exporter) Close() {
	e.file.Close()
}
