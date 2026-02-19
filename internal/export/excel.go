package export

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"ohnitiel/prismatic/internal/db"

	"github.com/xuri/excelize/v2"
)

// Styles are int because excelize.File.NewStyle() returns style index
type Styles struct {
	Number           int
	DateTime         int
	ConnectionColumn int
}

// Creates new default styles
func NewStyles(f *excelize.File) (*Styles, error) {
	dateStyle, err := f.NewStyle(&excelize.Style{
		NumFmt: 14,
	})
	if err != nil {
		return nil, err
	}

	decimalPlaces := 2
	numberStyle, err := f.NewStyle(&excelize.Style{
		NumFmt:        0,
		DecimalPlaces: &decimalPlaces,
	})
	if err != nil {
		return nil, err
	}

	connectionColumnStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	if err != nil {
		return nil, err
	}

	return &Styles{
		Number:           numberStyle,
		DateTime:         dateStyle,
		ConnectionColumn: connectionColumnStyle,
	}, nil
}

type ExcelOptions struct {
	SingleFile       bool
	SingleSheet      bool
	ConnectionColumn string
}

func NewExcelOptions(noSingleFile bool, noSingleSheet bool, connectionColumn string) ExcelOptions {
	return ExcelOptions{
		SingleFile:       noSingleFile,
		SingleSheet:      noSingleSheet,
		ConnectionColumn: connectionColumn,
	}
}

func Excel(
	ctx context.Context, data map[string]*db.ResultSet,
	output string, options ExcelOptions,
) error {
	switch {
	case options.SingleFile && !options.SingleSheet:
		return excelSingleFile(ctx, data, output)
	case !options.SingleFile && options.SingleSheet:
		return excelSingleSheet(ctx, data, output)
	default:
		return excelSingleFileAndSheet(ctx, data, output, options.ConnectionColumn)
	}
}

func excelSingleSheet(
	ctx context.Context, data map[string]*db.ResultSet,
	output string,
) error {
	sheetName := "Dados"
	for name, data := range data {
		f := excelize.NewFile()
		f.SetSheetName(f.GetSheetName(0), sheetName)
		colsWidth, err := writeDataToSheet(f, sheetName, data, "", "")
		if err != nil {
			slog.ErrorContext(ctx, "Error writing data to sheet", "error", err)
			return err
		}
		outputExt := filepath.Ext(output)
		output = strings.Replace(output, outputExt, fmt.Sprintf("_%s%s", name, outputExt), 1)

		for i, colWidth := range colsWidth {
			colName, _ := excelize.ColumnNumberToName(i)
			f.SetColWidth(sheetName, colName, colName, colWidth)
		}
		freezeHeader(f, sheetName)

		err = f.SaveAs(output)
		if err != nil {
			slog.ErrorContext(ctx, "Error saving file", "error", err)
			return err
		}
	}

	return nil
}

func excelSingleFile(ctx context.Context, data map[string]*db.ResultSet, output string) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			slog.ErrorContext(ctx, "Error closing file", "error", err)
		}
	}()

	for name, data := range data {
		f.NewSheet(name)
		colsWidth, err := writeDataToSheet(f, name, data, "", "")

		for i, colWidth := range colsWidth {
			colName, _ := excelize.ColumnNumberToName(i)
			f.SetColWidth(name, colName, colName, colWidth)
		}

		freezeHeader(f, name)
		if err != nil {
			slog.ErrorContext(ctx, "Error writing data to sheet", "error", err)
			return err
		}
	}

	f.DeleteSheet("Sheet1")

	err := f.SaveAs(output)
	if err != nil {
		slog.ErrorContext(ctx, "Error saving file", "error", err)
		return err
	}

	return nil
}

func excelSingleFileAndSheet(
	ctx context.Context, data map[string]*db.ResultSet,
	output string, connectionColumnName string,
) error {
	sheetName := "Dados"
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			slog.ErrorContext(ctx, "Error closing file", "error", err)
		}
	}()

	f.SetSheetName(f.GetSheetName(0), sheetName)
	f.SetActiveSheet(0)

	globalWidths := make(map[int]float64)

	for name, data := range data {
		colsWidth, err := writeDataToSheet(f, sheetName, data, connectionColumnName, name)
		if err != nil {
			slog.ErrorContext(ctx, "Error writing data to sheet", "error", err)
			return err
		}

		for idx, width := range colsWidth {
			if globalWidths[idx] < width {
				globalWidths[idx] = width
			}
		}
	}

	for idx, width := range globalWidths {
		colName, _ := excelize.ColumnNumberToName(idx)
		f.SetColWidth(sheetName, colName, colName, width)
	}

	freezeHeader(f, sheetName)
	err := f.SaveAs(output)
	if err != nil {
		slog.ErrorContext(ctx, "Error saving file", "error", err)
		return err
	}

	return nil
}

func writeDataToSheet(
	f *excelize.File, sheetName string,
	data *db.ResultSet, connectionName string, connectionColumn string,
) (map[int]float64, error) {
	if data.RowCount == 0 {
		return nil, fmt.Errorf("no data found")
	}

	sw, err := f.NewStreamWriter(sheetName)
	if err != nil {
		return nil, err
	}

	styles, err := NewStyles(f)
	if err != nil {
		return nil, err
	}

	if connectionName != "" {
		data.Columns = append(data.Columns, db.Column{
			Ordinal:  len(data.Columns),
			Name:     connectionName,
			Type:     "string",
			Nullable: false,
		})
	}

	columns := make([]db.Column, 0, len(data.Columns))
	for _, v := range data.Columns {
		columns = append(columns, v)
	}

	headers := make([]any, len(columns))
	for k, v := range columns {
		headers[k] = v.Name
	}

	sw.SetRow("A1", headers)

	colStyles := make(map[int]int, len(columns))
	for k, v := range columns {
		switch v.Type {
		case "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
			colStyles[k] = styles.Number
		case "Time":
			colStyles[k] = styles.DateTime
		}

		if v.Name == connectionName {
			colStyles[k] = styles.ConnectionColumn
		}
	}

	colsWidth := make(map[int]float64, len(columns))
	for i, row := range data.Rows {
		rowData := make([]any, len(headers))

		for j := range columns {
			var val any
			if columns[j].Name == connectionName {
				val = connectionColumn
			} else {
				val = row[j]
			}

			if styleID, ok := colStyles[j]; ok {
				rowData[j] = excelize.Cell{
					Value:   val,
					StyleID: styleID,
				}
			} else {
				rowData[j] = val
			}

			colsWidth[j] = max(colsWidth[j], float64(len(fmt.Sprintf("%v", val))))
		}

		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		sw.SetRow(cell, rowData)
	}

	lastCell, _ := excelize.CoordinatesToCellName(len(columns), data.RowCount+1)

	enabled := true
	err = sw.AddTable(&excelize.Table{
		Range:             fmt.Sprintf("A1:%s", lastCell),
		Name:              fmt.Sprintf("Tabela_%s", sheetName),
		StyleName:         "TableStyleMedium2",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowRowStripes:    &enabled,
		ShowColumnStripes: false,
	})
	if err != nil {
		return nil, err
	}

	return colsWidth, sw.Flush()
}

func freezeHeader(f *excelize.File, sheetName string) {
	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomRight",
	})
}
