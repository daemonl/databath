package sync

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/daemonl/databath"
	"github.com/daemonl/go_lib/extdb"
)

type TableStatus struct {
	Name            string  `db:"Name"`
	Engine          string  `db:"Engine"`
	Version         *uint64 `db:"Version"`
	Row_format      *string `db:"Row_format"`
	Rows            *uint64 `db:"Rows"`
	Avg_row_length  *uint64 `db:"Avg_row_length" json:"-"`
	Data_length     *uint64 `db:"Data_length" json:"-"`
	Max_data_length *uint64 `db:"Max_data_length" json:"-"`
	Index_length    *uint64 `db:"Index_length" json:"-"`
	Data_free       *uint64 `db:"Data_free" json:"-"`
	Auto_increment  *uint64 `db:"Auto_increment"`
	Create_time     *string `db:"Create_time" json:"-"`
	Update_time     *string `db:"Update_time" json:"-"`
	Check_time      *string `db:"Check_time" json:"-"`
	Collation       *string `db:"Collation"`
	Checksum        *string `db:"Checksum" json:"-"`
	Create_options  *string `db:"Create_options"`
	Comment         *string `db:"Comment"`
}

type ColumnStatus struct {
	Field   string  `db:"Field"`
	Type    string  `db:"Type"`
	Null    string  `db:"Null"`
	Key     *string `db:"Key"`
	Default *string `db:"Default"`
	Extra   *string `db:"Extra"`
}

type Index struct {
	ConstraintName       string  `db:"CONSTRAINT_NAME"`
	TableName            *string `db:"TABLE_NAME"`
	ConstraintType       *string `db:"CONSTRAINT_TYPE"`
	ColumnName           *string `db:"COLUMN_NAME"`
	ReferencedTableName  *string `db:"REFERENCED_TABLE_NAME"`
	ReferencedColumnName *string `db:"REFERENCED_COLUMN_NAME"`
	OwnerTable           *Table
	OwnedTable           *Table
	Used                 bool
}

type SyncError struct {
	Message string
}

func (e *SyncError) Error() string {
	return e.Message
}

var execString string = ""
var unused []string = []string{}

var reMidWhitespace *regexp.Regexp = regexp.MustCompile(`[\n\t\ ]+`)
var reLeadingWhitespace *regexp.Regexp = regexp.MustCompile(`^[\n\t\ ]+`)
var reTrailingWhitespace *regexp.Regexp = regexp.MustCompile(`[\n\t\ ]+$`)

var reCheckLength *regexp.Regexp = regexp.MustCompile(`^VARCHAR\(([0-9]+)\)`)

func (c *ColumnStatus) GetString() string {
	built := c.Type
	if c.Null == "NO" {
		built += " NOT NULL"
	} else {
		built += " NULL"
	}
	if c.Extra != nil {
		built += " " + *c.Extra
	}
	built = strings.TrimSpace(built)
	return strings.ToUpper(built)
}

func BuildMigration(db *sql.DB, model *databath.Model) (*Migration, error) {

	mig := &Migration{
		Checks:        []*Statement{},
		Statements:    []*Statement{},
		UnusedTables:  []string{},
		UnusedColumns: []string{},
	}

	edb := extdb.WrapDB(db)

	// Fix any non InnoDB tables

	nonInno := []*TableStatus{}
	err := edb.Select(&nonInno, `SHOW TABLE STATUS WHERE Engine != 'InnoDB'`)
	if err != nil {
		return nil, fmt.Errorf("Loading Table != InnoDB Status: %s", err.Error())
	}

	for _, table := range nonInno {
		s := Statementf("ALTER TABLE %s ENGINE = 'InnoDB'", table.Name)
		s.Owner = table.Name
		mig.Statements = append(mig.Statements, s)
	}

	// Load the table info to memory.
	tableStatuses := []*TableStatus{}
	tables := map[string]*Table{}
	if err = edb.Select(&tableStatuses, `SHOW TABLE STATUS WHERE ENGINE IS NOT NULL`); err != nil {
		return nil, fmt.Errorf("Loading all table status: %s", err.Error())
	}

	for _, tableStatus := range tableStatuses {
		t := getBlankTable(tableStatus.Name)

		t.Status = tableStatus
		tables[tableStatus.Name] = t
		constraints := []*Index{}
		if err = edb.Select(&constraints, `
			SELECT
			  c.CONSTRAINT_NAME,
			  c.TABLE_NAME,
			  c.CONSTRAINT_TYPE,
			  k.COLUMN_NAME,
			  k.REFERENCED_TABLE_NAME,
			  k.REFERENCED_COLUMN_NAME 
			FROM information_schema.TABLE_CONSTRAINTS c
			LEFT JOIN information_schema.KEY_COLUMN_USAGE k 
			  ON c.CONSTRAINT_NAME = k.CONSTRAINT_NAME 
			  AND c.TABLE_SCHEMA = k.TABLE_SCHEMA 
			  AND c.TABLE_NAME = k.TABLE_NAME 
			WHERE c.TABLE_SCHEMA = DATABASE() AND c.TABLE_NAME = ?`, tableStatus.Name); err != nil {
			return nil, fmt.Errorf("Error selecting constraints for %s: %s", tableStatus.Name, err.Error())
		}

		for _, constraint := range constraints {
			t.Indexes[constraint.ConstraintName] = constraint
		}

		columnStatuses := []*ColumnStatus{}
		if err = edb.Select(&columnStatuses, `SHOW COLUMNS FROM `+tableStatus.Name); err != nil { // Can't use '?'
			return nil, fmt.Errorf("Error selecting columns for %s: %s", tableStatus.Name, err.Error())
		}
		for _, columnStatus := range columnStatuses {
			column := &Column{
				Status: columnStatus,
				Name:   columnStatus.Field,
				Table:  t,
			}
			t.Columns[columnStatus.Field] = column
		}
	}

	// Create and sync all the collections
	for collectionName, collection := range model.Collections {

		table, ok := tables[collectionName]
		if !ok {
			table = getBlankTable(collectionName)
			tables[collectionName] = table
		}
		table.Collection = collection

		for fieldName, field := range collection.Fields {
			column, ok := table.Columns[fieldName]
			if !ok {
				column = &Column{
					Table: table,
					Name:  fieldName,
				}
				table.Columns[fieldName] = column
			}
			column.Field = field
		}
	}

	for _, t := range tables {
		err := t.Sync()
		if err != nil {
			return nil, err
		}
		err = t.setupIndexes()
		if err != nil {
			return nil, fmt.Errorf("Error seting indexes on table %s: %s", t.Name, err.Error())
		}
	}

	for _, t := range tables {
		if t.Collection == nil {
			mig.UnusedTables = append(mig.UnusedTables, t.Name)
			continue
		}
		for _, check := range t.Checks {
			mig.Checks = append(mig.Checks, check)
		}
		for _, s := range t.Statements {
			mig.Statements = append(mig.Statements, s)
		}
		for _, col := range t.Columns {
			if col.Field == nil {
				mig.UnusedColumns = append(mig.UnusedColumns, t.Name+"."+col.Name)
			}
		}
	}

	for _, t := range tables {
		for _, s := range t.PostStatements {
			mig.Statements = append(mig.Statements, s)
		}
	}

	return mig, nil

}
