package db

import "time"

type Column struct {
	Ordinal  int
	Name     string
	Type     string
	Nullable bool
}

type ResultSet struct {
	Columns  []Column
	Rows     [][]any
	RowCount int
	Duration time.Duration
}
