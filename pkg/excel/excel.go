// Package excel 提供 Excel 导出能力。
//
// 设计要点（相对早前版本的修正）：
//   - 所有错误显式返回。早前 `SetCellValue` 的返回值被整体丢弃，
//     sheet 不存在、坐标非法等问题会静默丢数据，导出结果缺列却毫无提示
//   - sheet 名与行号由 Exporter 自己维护，调用方不再手工传行号。
//     早前 SetHeaders/SetRow 要求调用方传 sheet 名与行号，一旦传错
//     （例如对着不存在的 sheet 写入）就是静默失败
//   - 构造时直接指定 sheet 名并重命名默认表，避免导出的文件里
//     残留一个空白的 Sheet1
package excel

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

// Exporter 按「表头 + 逐行追加」的方式构建 Excel
type Exporter struct {
	file  *excelize.File
	sheet string
	row   int // 下一个可写行号（1 起）
}

// NewExporter 创建导出器。sheetName 为空时使用 excelize 默认的 "Sheet1"。
func NewExporter(sheetName string) (*Exporter, error) {
	file := excelize.NewFile()

	defaultSheet := file.GetSheetName(0)
	if sheetName == "" {
		sheetName = defaultSheet
	}

	// 重命名默认表而不是新增一张，否则导出的文件里会多出一个空白 sheet
	if sheetName != defaultSheet {
		if err := file.SetSheetName(defaultSheet, sheetName); err != nil {
			_ = file.Close()
			return nil, fmt.Errorf("设置 sheet 名失败: %w", err)
		}
	}

	return &Exporter{file: file, sheet: sheetName, row: 1}, nil
}

// SetHeaders 写入表头，占用第 1 行
func (e *Exporter) SetHeaders(headers []string) error {
	if len(headers) == 0 {
		return nil
	}
	if err := e.writeRow(1, toStrings(headers)); err != nil {
		return fmt.Errorf("写入表头失败: %w", err)
	}
	e.row = 2
	return nil
}

// AddRow 追加一行数据，行号由导出器自动递增
func (e *Exporter) AddRow(values ...interface{}) error {
	if err := e.writeRow(e.row, values); err != nil {
		return fmt.Errorf("写入第 %d 行失败: %w", e.row, err)
	}
	e.row++
	return nil
}

// RowCount 返回已写入的数据行数（不含表头）
func (e *Exporter) RowCount() int {
	if e.row <= 1 {
		return 0
	}
	return e.row - 2
}

func (e *Exporter) writeRow(row int, values []interface{}) error {
	for i, v := range values {
		cell, err := excelize.CoordinatesToCellName(i+1, row)
		if err != nil {
			return fmt.Errorf("计算单元格坐标失败(列 %d, 行 %d): %w", i+1, row, err)
		}
		if err := e.file.SetCellValue(e.sheet, cell, v); err != nil {
			return err
		}
	}
	return nil
}

// WriteToWriter 把工作簿写出到 w，适用于直接回写给 HTTP 响应。
//
// 刻意不叫 WriteTo：那会与 io.WriterTo 的标准签名
// WriteTo(io.Writer) (int64, error) 冲突，go vet 会告警，
// 也容易让调用方误以为实现了该接口。
func (e *Exporter) WriteToWriter(w io.Writer) error {
	return e.file.Write(w)
}

// SaveAs 保存到本地文件
func (e *Exporter) SaveAs(path string) error {
	return e.file.SaveAs(path)
}

// Close 释放底层资源。导出完成后必须调用。
func (e *Exporter) Close() error {
	return e.file.Close()
}

func toStrings(in []string) []interface{} {
	out := make([]interface{}, len(in))
	for i, s := range in {
		out[i] = s
	}
	return out
}
