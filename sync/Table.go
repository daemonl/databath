package sync

import (
	"fmt"
	"strings"

	"github.com/daemonl/databath"
)

type Table struct {
	Name           string
	Status         *TableStatus
	Columns        map[string]*Column
	Indexes        map[string]*Index
	Collection     *databath.Collection
	Statements     []*Statement
	PostStatements []*Statement
	Checks         []*Statement
}

func getBlankTable(name string) *Table {
	return &Table{
		Name:           name,
		Indexes:        map[string]*Index{},
		Columns:        map[string]*Column{},
		Statements:     []*Statement{},
		PostStatements: []*Statement{},
		Checks:         []*Statement{},
	}
}

type RefField interface {
	GetRefCollectionName() string
}

func (t *Table) addStatement(s *Statement) {
	t.Statements = append(t.Statements, s)
}

func (t *Table) addCheck(s *Statement) {
	t.Checks = append(t.Checks, s)
}

func (t *Table) addPost(s *Statement) {
	t.PostStatements = append(t.PostStatements, s)
}

/*
func (t *Table) addStatementf(s string, p ...interface{}) {
	t.Statements = append(t.Statements, fmt.Sprintf(s, p...))
}

func (t *Table) addPostStatementf(s string, p ...interface{}) {
	t.PostStatements = append(t.PostStatements, fmt.Sprintf(s, p...))
}

func (t *Table) addCheckf(s string, p ...interface{}) {
	t.Checks = append(t.Checks, fmt.Sprintf(s, p...))
}
*/

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
			s := Statementf("ALTER TABLE %s DROP FOREIGN KEY %s", t.Name, index.ConstraintName)
			s.Owner = t.Name + ":FK:" + index.ConstraintName
			s.Notes = "Unused Foreign Key"
			t.addStatement(s)
		}
	}
	return nil
}

func (t *Table) setupViewQuery() error {
	//TODO: Destroy any foreign keys.
	sDrop := Statementf("DROP TABLE IF EXISTS %s", t.Name)
	sCreate := Statementf("CREATE OR REPLACE VIEW %s AS %s", t.Name, *t.Collection.ViewQuery)
	sDrop.Owner = t.Name
	sCreate.Owner = t.Name
	sDrop.Notes = "Just in case"
	sCreate.Notes = "Always run, not only on changes"
	t.addStatement(sDrop)
	t.addStatement(sCreate)
	return nil
}

func (t *Table) create() error {
	params := make([]string, 0, 0)

	for name, field := range t.Collection.Fields {
		params = append(params, fmt.Sprintf("`%s` %s", name, field.GetMysqlDef()))
	}

	params = append(params, "PRIMARY KEY (`id`)")

	s := Statementf("CREATE TABLE %s (%s)", t.Name, strings.Join(params, ", "))
	s.Owner = t.Name
	t.addStatement(s)

	return nil
}
