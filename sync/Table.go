package sync

import (
	"fmt"
	"log"
	"strings"

	"github.com/daemonl/databath"
)

type Table struct {
	Name           string
	Status         *TableStatus
	Columns        map[string]*Column
	Indexes        map[string]*Index
	Collection     *databath.Collection
	Statements     []string
	PostStatements []string
	Checks         []string
}

type RefField interface {
	GetRefCollectionName() string
}

func (t *Table) addStatementf(s string, p ...interface{}) {
	t.Statements = append(t.Statements, fmt.Sprintf(s, p...))
}

func (t *Table) addPostStatementf(s string, p ...interface{}) {
	t.PostStatements = append(t.PostStatements, fmt.Sprintf(s, p...))
}

func (t *Table) addCheckf(s string, p ...interface{}) {
	t.Checks = append(t.Checks, fmt.Sprintf(s, p...))
}

func (t *Table) Sync() error {

	// Should the table exist?
	if t.Collection == nil {
		return nil
	}

	// View query specifies a view syntax instead of a table.
	if t.Collection.ViewQuery != nil {
		err := t.setupViewQuery()
		return err
	}

	// Does the table exist?
	if t.Status == nil {
		// Create a table.
		t.create()
		return nil
	}

	for _, c := range t.Columns {
		err := c.Sync()
		if err != nil {
			return fmt.Errorf("on %s.%s: %s", t.Name, c.Name, err.Error())
		}
	}

	return nil
}

func (t *Table) setupIndexes() error {

	for _, col := range t.Columns {
		err := col.doIndexes()
		if err != nil {
			return err
		}
	}

	for _, index := range t.Indexes {
		if !index.Used && *index.ConstraintType == "FOREIGN KEY" {
			t.addStatementf("ALTER TABLE %s DROP FOREIGN KEY %s", t.Name, index.ConstraintName)
		}
	}
	return nil
}

func (t *Table) setupViewQuery() error {
	//TODO: Destroy any foreign keys.
	MustExecF(now, db, "DROP TABLE IF EXISTS %s", collectionName)
	MustExecF(now, db, "CREATE OR REPLACE VIEW %s AS %s", collectionName, *collection.ViewQuery)
	log.Println("SKIP COLLECTION - It has a view query")
	return nil
}

func (t *Table) create() error {
	params := make([]string, 0, 0)

	for name, field := range t.Collection.Fields {
		params = append(params, fmt.Sprintf("`%s` %s", name, field.GetMysqlDef()))
	}

	params = append(params, "PRIMARY KEY (`id`)")

	t.addStatementf("CREATE TABLE %s (%s)", t.Name, strings.Join(params, ", "))

	return nil
}
