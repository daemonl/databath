package model

import (
	"strings"

	"github.com/daemonl/databath/types"
)

type Field struct {
	types.FieldType
	Raw        map[string]interface{}
	OnCreate   *interface{}
	Path       string
	Collection *Collection
}

func (f *Field) Init(raw map[string]interface{}) error {
	f.Raw = raw
	onCreate, ok := raw["on_create"]
	if ok {
		f.OnCreate = &onCreate
	}
	//err := f.Impl.Init(raw)
	return nil
}

type Context interface {
	GetValueFor(string) interface{}
}

func (f *Field) GetDefault(context Context) (string, error) {
	if f.OnCreate == nil {
		return "", nil
	}

	strVal, isStr := (*f.OnCreate).(string)
	if isStr && strings.HasPrefix(strVal, "#") {
		val := context.GetValueFor(strVal[1:])
		//fmt.Sprintf("##################%s  %s  %s\n\n\n", strVal, strVal[1:], val)
		return f.ToDb(val)
	}
	return f.ToDb(*f.OnCreate)
}

type FieldType interface {
	//Init(map[string]interface{}) error
	FromDb(interface{}) (interface{}, error)
	ToDb(interface{}) (string, error)
	GetScanReciever() interface{}
	IsSearchable() bool
	GetMysqlDef() string
}

/*
func FieldFromDef(rawField map[string]interface{}) (*Field, error) {

	// field must have type
	fieldType, err := getFieldParamString(rawField, "type")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s", err.Error()))
	}
	if fieldType == nil {
		return nil, errors.New(fmt.Sprintf("no type specified"))
	}

	typeString := *fieldType
	typeString = strings.Replace(typeString, "_", "", -1)
	typeString = strings.ToLower(typeString)

	fieldImpl, err := FieldByType(FieldTypeName(typeString))
	if err != nil {
		return nil, err
	}
	field := &Field{Impl: fieldImpl}
	//err = field.Init(rawField)
	//if err != nil {
	//	return nil, err
	//}

	return field, nil
}*/
