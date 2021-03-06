package databath

import (
	"errors"
	"fmt"
	"strings"

	"github.com/daemonl/databath/types"
)

type Field struct {
	Impl       FieldType
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
	err := f.Impl.Init(raw)
	return err
}

func (f *Field) FromDb(raw interface{}) (interface{}, error) { return f.Impl.FromDb(raw) }
func (f *Field) ToDb(raw interface{}, context Context) (string, error) {
	/*
		strVal, ok := raw.(string)
		if ok {
			if strings.HasPrefix(strVal, "#") {
				raw = context.getValueFor(strVal)
			}
		}*/

	return f.Impl.ToDb(raw)
}

func (f *Field) GetScanReciever() interface{} { return f.Impl.GetScanReciever() }
func (f *Field) IsSearchable() bool           { return f.Impl.IsSearchable() }
func (f *Field) GetMysqlDef() string          { return f.Impl.GetMysqlDef() }

func (f *Field) GetDefault(context Context) (string, error) {
	if f.OnCreate == nil {
		return "", nil
	}

	strVal, isStr := (*f.OnCreate).(string)
	if isStr && strings.HasPrefix(strVal, "#") {
		val := context.GetValueFor(strVal[1:])
		//fmt.Sprintf("##################%s  %s  %s\n\n\n", strVal, strVal[1:], val)
		return f.Impl.ToDb(val)
	}
	return f.Impl.ToDb(*f.OnCreate)
}

type FieldType interface {
	Init(map[string]interface{}) error
	FromDb(interface{}) (interface{}, error)
	ToDb(interface{}) (string, error)
	GetScanReciever() interface{}
	IsSearchable() bool
	GetMysqlDef() string
}

func FieldByType(typeString string) (FieldType, error) {

	typeString = strings.Replace(typeString, "_", "", -1)
	typeString = strings.ToLower(typeString)

	switch typeString {
	case "string":
		return &types.FieldString{}, nil
	case "id":
		return &types.FieldId{}, nil
	case "ref":
		return &types.FieldRef{}, nil
	case "array":
		return &types.FieldString{}, nil
	case "datetime":
		return &types.FieldDateTime{}, nil
	case "date":
		return &types.FieldDate{}, nil
	case "time":
		return &types.FieldTime{}, nil
	case "int":
		return &types.FieldInt{}, nil
	case "range":
		return &types.FieldInt{}, nil
	case "bool":
		return &types.FieldBool{}, nil
	case "text":
		return &types.FieldText{}, nil
	case "address":
		return &types.FieldText{}, nil
	case "float":
		return &types.FieldFloat{}, nil
	case "password":
		return &types.FieldPassword{}, nil
	case "file":
		return &types.FieldFile{}, nil
	case "enum":
		return &types.FieldEnum{}, nil
	case "patientcard":
		return &types.FieldPatientCard{}, nil

	case "color":
		return &types.FieldString{}, nil

	case "autotimestamp":
		return &types.FieldInt{}, nil

	case "timestamp":
		return &types.FieldTimestamp{}, nil
	case "sqltimestamp":
		return &types.FieldTimestamp{}, nil

	case "???":
		return &types.FieldString{}, nil

	case "bitswitch":
		return &types.FieldInt{}, nil

	case "keyval":
		return &types.FieldKeyVal{}, nil

	case "blobject":
		return &types.FieldBlobject{}, nil

	case "gob":
		return &types.FieldGob{}, nil

	case "refid":
		return &types.FieldRefID{}, nil

	default:
		return nil, errors.New(fmt.Sprintf("Invalid Field Type '%s'", typeString))
	}
}

func FieldFromDef(rawField map[string]interface{}) (*Field, error) {

	// field must have type
	fieldType, err := getFieldParamString(rawField, "type")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s", err.Error()))
	}
	if fieldType == nil {
		return nil, errors.New(fmt.Sprintf("no type specified"))
	}
	fieldImpl, err := FieldByType(*fieldType)
	if err != nil {
		return nil, err
	}
	field := &Field{Impl: fieldImpl}
	err = field.Init(rawField)
	if err != nil {
		return nil, err
	}

	return field, nil
}
