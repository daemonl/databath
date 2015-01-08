package sync

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type Migration struct {
	Checks        []*Statement
	Statements    []*Statement
	UnusedTables  []string
	UnusedColumns []string
}

func (mig *Migration) Run(db *sql.DB) error {
	checksFailed := false
	for _, check := range mig.Checks {
		log.Printf("%s: %s\n", check.Owner, check.SQL)
		rows, err := db.Query(check.SQL)
		if err != nil {
			return err
		}
		columns, _ := rows.Columns()
		for rows.Next() {
			row := make([]interface{}, len(columns))
			strs := make([]string, len(columns))
			for i := range row {
				row[i] = &strs[i]
			}
			err := rows.Scan(row...)
			if err != nil {
				return err
			}
			fmt.Println(strings.Join(strs, ", "))
			checksFailed = true
		}
		rows.Close()

	}

	if checksFailed {
		return fmt.Errorf("Checks Failed")
	}

	for _, statement := range mig.Statements {
		log.Printf("%s: %s\n", statement.Owner, statement.SQL)
		_, err := db.Exec(statement.SQL)
		if err != nil {
			return err
		}
	}

	return nil
}
