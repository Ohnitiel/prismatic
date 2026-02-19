package sql

import (
	"fmt"
	"strings"
)

type QueryType int

const (
	DQL QueryType = iota
	DML
	DDL
)

func (qt QueryType) String() string {
	return []string{"DQL", "DML", "DDL"}[qt]
}

func (qt QueryType) IsSafe() bool {
	return qt == DQL
}

// Very dumb query type identifier
// TODO: Add a parser and use that
func SimpleQueryIdentifier(query string) (QueryType, error) {
	if strings.Contains(strings.ToUpper(query), "SELECT") {
		return DQL, nil
	} else if strings.Contains(strings.ToUpper(query), "INSERT") {
		return DML, nil
	} else if strings.Contains(strings.ToUpper(query), "UPDATE") {
		return DML, nil
	} else if strings.Contains(strings.ToUpper(query), "DELETE") {
		return DML, nil
	} else if strings.Contains(strings.ToUpper(query), "CREATE") {
		return DDL, nil
	} else if strings.Contains(strings.ToUpper(query), "DROP") {
		return DDL, nil
	}

	return 0, fmt.Errorf("Unable to identify query type")
}
