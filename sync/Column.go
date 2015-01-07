package sync

import (
	"strconv"

	"github.com/daemonl/databath"
)

type Column struct {
	Name   string
	Table  *Table
	Field  *databath.Field
	Status *ColumnStatus
}

func (c *Column) Sync() error {
	// Edge conditions?
	if c.Field == nil {
		return
	}
	if c.Status == nil {
		// Create new
		c.Table.addStatementf("ALTER TABLE %s ADD `%s` %s", c.Table.Name, c.Name, c.Field.GetMysqlDef())
		return nil
	}

	colStr := c.Status.GetString()
	modelStr := c.Field.GetMysqlDef()
	if colStr == modelStr {
		// Table matches, no issues.
		return nil
	}

	// If VARCHAR(100) etc
	if reCheckLength.MatchString(modelStr) {
		matches := reCheckLength.FindStringSubmatch(modelStr)
		lenNewMax, _ := strconv.ParseUint(matches[1], 10, 64)
		// TODO: READ THE THINGS!
		c.Table.addCheckf("SELECT id FROM %s WHERE LENGTH(%s) > %d")
	}

	c.Table.addStatementf("ALTER TABLE %s CHANGE COLUMN %s %s %s",
		collectionName, colName, colName, modelStr)

	// TODO: LENGTH CHECKS!
	return nil

}

func (c *Column) setupIndexes() error {
	return c.doRefField()
}

func (c *Column) doRefField() error {
	refField, ok := c.Field.Impl.(RefField)
	if !ok {
		return nil
	}
	linkTo := refField.GetRefCollectionName()

	matchingIndex, ok := c.findIndexMatching(c.Name, linkTo)
	if ok {
		matchingIndex.Used == true
		// TODO: Check and update the index?
	} else {
		c.Table.addCheckf(`
		SELECT id, %s 
		FROM %s 
		WHERE %s IS NOT NULL 
		AND %s NOT IN 
		(SELECT id FROM %s)`,
			c.Name, c.Table.Name, c.Name, c.Name, linkTo)

		c.Table.addPostStatementf(`
		ALTER TABLE %s 
		ADD CONSTRAINT fk_%s_%s 
		FOREIGN KEY (%s) 
		REFERENCES %s(id)`,
			c.Table.Name, c.Table.Name, c.Name, c.Name, linkTo)
	}

}

func (t *Table) findIndexMatching(fieldName string, collectionName string) (*Index, bool) {

	for _, index := range t.Indexes {
		if index.ReferencedTableName == nil || index.ColumnName == nil {
			continue
		}
		if *index.ColumnName == fieldName && *index.ReferencedTableName == collectionName {
			return index, true
		}
	}
	return nil, false
}
