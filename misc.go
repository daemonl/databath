package databath

import (
	"fmt"
	"net/http"

	"github.com/daemonl/databath/types"
)

type DynamicFunction struct {
	Filename   string   `json:"filename"`
	Parameters []string `json:"parameters"`
	Access     []uint64 `json:"access"`
}

type MappedTable struct {
	path       string
	alias      string
	collection *Collection
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
		if _, ok := mf.field.Impl.(*types.FieldBlobject); ok {
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

type Context interface {
	GetValueFor(string) interface{}
	GetUserLevel() (isApplication bool, userAccessLevel uint64)
}

type MapContext struct {
	IsApplication   bool
	UserAccessLevel uint64
	Fields          map[string]interface{}
}

func (mc *MapContext) GetUserLevel() (bool, uint64) {
	return mc.IsApplication, mc.UserAccessLevel
}

func (mc *MapContext) GetValueFor(key string) interface{} {
	val, ok := mc.Fields[key]
	if !ok {
		return key
	}
	return val
}

type QueryUserError struct {
	Message  string
	HTTPCode int
}

func (ue *QueryUserError) Error() string {
	return ue.Message
}

func UserErrorF(format string, params ...interface{}) *QueryUserError {
	return &QueryUserError{Message: fmt.Sprintf(format, params...), HTTPCode: http.StatusBadRequest}
}

func UserAccessErrorF(format string, params ...interface{}) *QueryUserError {
	return &QueryUserError{Message: fmt.Sprintf(format, params...), HTTPCode: http.StatusForbidden}
}

func (ue *QueryUserError) GetUserDescription() string {
	return ue.Message
}
func (ue *QueryUserError) GetHTTPStatus() int {
	return ue.HTTPCode
}
