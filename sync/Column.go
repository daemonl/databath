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
		return nil
	}
	if c.Status == nil {
		// Create new
		s := Statementf("ALTER TABLE %s ADD `%s` %s", c.Table.Name, c.Name, c.Field.GetMysqlDef())
		s.Owner = c.Table.Name + "." + c.Name
		c.Table.addStatement(s)
		return nil
	}

	colStr := c.Status.GetString()
	modelStr := c.Field.GetMysqlDef()
	if colStr == modelStr {
		// Table matches, no issues.
		return nil
	}
	if colStr == "TIMESTAMP NOT NULL" &&
		modelStr == "TIMESTAMP DEFAULT CURRENT_TIMESTAMP" {
		return nil
	}

	// If VARCHAR(100) etc
	if reCheckLength.MatchString(modelStr) {
		matches := reCheckLength.FindStringSubmatch(modelStr)
		lenNewMax, _ := strconv.ParseUint(matches[1], 10, 64)
		// TODO: READ THE THINGS!
		s := Statementf("SELECT id FROM %s WHERE LENGTH(%s) > %d", c.Table.Name, c.Name, lenNewMax)
		s.Owner = c.Table.Name + "." + c.Name
		c.Table.addCheck(s)
	}

	s := Statementf("ALTER TABLE %s CHANGE COLUMN `%s` `%s` %s",
		c.Table.Name, c.Name, c.Name, modelStr)
	s.Owner = c.Table.Name + "." + c.Name
	s.Notes = "Existing: " + colStr
	c.Table.addStatement(s)

	// TODO: LENGTH CHECKS!
	return nil

}

func (c *Column) doIndexes() error {
	return c.doRefField()
}

func (c *Column) doRefField() error {
	if c.Field == nil {
		return nil
	}
	refField, ok := c.Field.FieldType.(RefField)
	if !ok {
		return nil
	}
	linkTo := refField.GetRefCollectionName()

	matchingIndex, ok := c.Table.findIndexMatching(c.Name, linkTo)
	if ok {
		matchingIndex.Used = true
		// TODO: Check and update the index?
	} else {
		sCheck := Statementf(`
		SELECT id, %s 
		FROM %s 
		WHERE %s IS NOT NULL 
		AND %s NOT IN 
		(SELECT id FROM %s)`,
			c.Name, c.Table.Name, c.Name, c.Name, linkTo)

		sPost := Statementf(`
		ALTER TABLE %s 
		ADD CONSTRAINT fk_%s_%s 
		FOREIGN KEY (%s) 
		REFERENCES %s(id)`,
			c.Table.Name, c.Table.Name, c.Name, c.Name, linkTo)

		sCheck.Owner = c.Table.Name + "." + c.Name
		sPost.Owner = c.Table.Name + "." + c.Name
		c.Table.addCheck(sCheck)
		c.Table.addPost(sPost)
	}
	return nil

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
