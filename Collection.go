package databath

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/daemonl/databath/types"
)

type Collection struct {
	Model          *Model
	Fields         map[string]*Field
	FieldSets      map[string][]FieldSetFieldDef
	CustomFields   map[string]FieldSetFieldDef
	Masks          map[uint64]*Mask
	ForeignKeys    []*Field
	Hooks          []IHook
	TableName      string
	SearchPrefixes map[string]*SearchPrefix
	ViewQuery      *string
}

type SearchPrefix struct {
	Prefix    string
	Field     Field
	FieldName string
}

func (c *Collection) AddHook(hook IHook) {
	c.Hooks = append(c.Hooks, hook)
}

func (c *Collection) GetFieldSet(fieldSetNamePointer *string) ([]FieldSetFieldDef, error) {
	var fieldSetName string
	if fieldSetNamePointer == nil {
		fieldSetName = "default"
	} else {
		fieldSetName = *fieldSetNamePointer
	}

	fields, ok := c.FieldSets[fieldSetName]
	if !ok {
		return nil, UserErrorF("Fieldset %s doesn't exist in %s", fieldSetName, c.TableName)
	}
	log.Printf("Using fieldset: %s.%s\n", c.TableName, fieldSetName)

	return fields, nil
}

func (c *Collection) CheckDelete(db *sql.DB, id uint64) (*DeleteCheckResult, error) {

	dd := &DeleteCheckResult{
		Children:           make(map[string]map[string]*DeleteCheckResult),
		ToExecute:          make([]string, 0, 0),
		WillBeDeleted:      make(map[string][]uint64),
		PreventingDeletion: make(map[string][]uint64),
		Prevents:           false,
	}

	for _, field := range c.ForeignKeys {
		refField, ok := field.Impl.(*types.FieldRef)
		if !ok {
			return nil, ParseErrF("Foreign key not a ref type")
		}

		sql := fmt.Sprintf("SELECT id FROM `%s` WHERE  %s.%s = %d", field.Collection.TableName, field.Collection.TableName, field.Path, id)
		log.Printf("CHECK DELETE: %s\n", sql)
		res, err := db.Query(sql)
		if err != nil {
			return nil, err
		}
		defer res.Close()
		existingRefs := make([]uint64, 0, 0)
		for res.Next() {
			var refId uint64
			res.Scan(&refId)
			existingRefs = append(existingRefs, refId)
			if refField.OnDelete == types.RefOnDeleteCascade {
				dd.WillBeDeleted[field.Collection.TableName] = existingRefs
				refDeleteCheckResult, err := c.Model.Collections[field.Collection.TableName].CheckDelete(db, refId)
				if err != nil {
					return nil, err
				}
				dd.ToExecute = append(dd.ToExecute, fmt.Sprintf("DELETE FROM `%s` WHERE id = %d", field.Collection.TableName, refId))
				dd.AddChild(refDeleteCheckResult, field.Collection.TableName, refId)

			} else if refField.OnDelete == types.RefOnDeleteNull {
				dd.ToExecute = append(dd.ToExecute, fmt.Sprintf("UPDATE %s SET %s = NULL WHERE id = %d", field.Collection.TableName, field.Path, refId))

			} else {
				dd.PreventingDeletion[field.Collection.TableName] = existingRefs
				dd.Prevents = true
			}
		}
	}

	return dd, nil
}
