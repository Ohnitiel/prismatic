package export

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
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

		sw, err := f.NewStreamWriter(sheetName)
		if err != nil {
			return err
		}

		styles, err := NewStyles(f)
		if err != nil {
			return err
		}

		colsWidth, err := writeDataToSheet(
			sw, styles, 1, sheetName, data, "", "", true,
		)
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

		err = sw.Flush()
		if err != nil {
			slog.ErrorContext(ctx, "Error flushing data to sheet", "error", err)
			return err
		}

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

	styles, err := NewStyles(f)
	if err != nil {
		return err
	}

	for name, data := range data {
		f.NewSheet(name)

		sw, err := f.NewStreamWriter(name)
		if err != nil {
			return err
		}

		colsWidth, err := writeDataToSheet(
			sw, styles, 1, name, data, "", "", true,
		)
		if err != nil {
			slog.ErrorContext(ctx, "Error writing data to sheet", "error", err)
			return err
		}

		err = sw.Flush()
		if err != nil {
			slog.ErrorContext(ctx, "Error flushing data to sheet", "error", err)
			return err
		}

		for i, colWidth := range colsWidth {
			colName, _ := excelize.ColumnNumberToName(i)
			f.SetColWidth(name, colName, colName, colWidth)
		}

		freezeHeader(f, name)

	}

	f.DeleteSheet("Sheet1")

	err = f.SaveAs(output)
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

	styles, err := NewStyles(f)
	if err != nil {
		return err
	}

	f.SetSheetName(f.GetSheetName(0), sheetName)
	f.SetActiveSheet(0)

	sw, err := f.NewStreamWriter(sheetName)
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	currRow := 1
	addTable := false
	globalWidths := make(map[int]float64)
	for k, name := range keys {
		if k == len(keys)-1 {
			addTable = true
		}
		result := data[name]
		colsWidth, err := writeDataToSheet(
			sw, styles, currRow, sheetName, result, name, connectionColumnName, addTable,
		)
		if err != nil {
			slog.ErrorContext(ctx, "Error writing data to sheet", "error", err)
			return err
		}

		for idx, width := range colsWidth {
			if globalWidths[idx] < width {
				globalWidths[idx] = width
			}
		}

		currRow += result.RowCount
	}

	err = sw.Flush()
	if err != nil {
		slog.ErrorContext(ctx, "Error flushing data to sheet", "error", err)
		return err
	}

	for idx, width := range globalWidths {
		colName, _ := excelize.ColumnNumberToName(idx)
		f.SetColWidth(sheetName, colName, colName, width)
	}

	freezeHeader(f, sheetName)
	err = f.SaveAs(output)
	if err != nil {
		slog.ErrorContext(ctx, "Error saving file", "error", err)
		return err
	}

	return nil
}

func writeDataToSheet(
	sw *excelize.StreamWriter, styles *Styles, startRow int,
	sheetName string, data *db.ResultSet,
	connectionName string, connectionColumn string, addTable bool,
) (map[int]float64, error) {
	if data.RowCount == 0 {
		return nil, fmt.Errorf("no data found")
	}

	if connectionName != "" {
		data.Columns = append(data.Columns, db.Column{
			Ordinal:  len(data.Columns),
			Name:     connectionColumn,
			Type:     "string",
			Nullable: false,
		})
	}

	columns := make([]db.Column, 0, len(data.Columns))
	for _, v := range data.Columns {
		columns = append(columns, v)
	}

	if startRow == 1 {

		headers := make([]any, len(columns))
		for k, v := range columns {
			headers[k] = v.Name
		}

		sw.SetRow("A1", headers)
	}

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
		rowData := make([]any, len(columns))

		for j := range columns {
			var val any
			if columns[j].Name == connectionColumn {
				val = connectionName
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

		cell, _ := excelize.CoordinatesToCellName(1, i+1+startRow)
		sw.SetRow(cell, rowData)
	}

	lastCell, _ := excelize.CoordinatesToCellName(len(columns), data.RowCount+startRow)

	if addTable {
		enabled := true
		err := sw.AddTable(&excelize.Table{
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
	}

	return colsWidth, nil
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
