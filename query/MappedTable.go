package query

import (
	"fmt"

	"github.com/daemonl/databath/types"
)

type MappedTable struct {
	path       string
	alias      string
	collection *model.Collection
}

type MappedField struct {
	path             string
	alias            string
	fieldNameInTable string
	fieldSetFieldDef FieldSetFieldDef
	field            *Field
	table            *MappedTable
	def              *Collection
	selectString     *string
	AllowSearch      bool
}

func (mf *MappedField) CanSearch() bool {
	if !mf.field.IsSearchable() {
		return false
	}
	return mf.AllowSearch
}

func (mf *MappedField) ConstructQuery(term string) *QueryConditionWhere {
	if mf.CanSearch() {
		if _, ok := mf.field.FieldType.(*types.FieldBlobject); ok {
			condition := QueryConditionWhere{
				Field: mf.path,
				Cmp:   "INJSON",
				Val:   term,
			}
			return &condition
		} else {
			condition := QueryConditionWhere{
				Field: mf.path,
				Cmp:   "LIKE",
				Val:   term,
			}
			return &condition
		}
	} else {
		fmt.Printf("Can't search mapped field %s\n", mf.path)
		return nil
	}
}
