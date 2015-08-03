package query_builder

import "database/sql"

type SelectQuery struct {
	columns    []QueryColumn
	parameters []interface{}
	sql        string
}

type QueryColumn interface{}

func (q *SelectQuery) Run(db *sql.DB) error {
	rows, err := db.Query(q.sql, q.parameters...)
	if err != nil {
		return err
	}
	_ = rows
	// do something with rows.
	return nil
}
