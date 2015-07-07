package types

import (
	"errors"
	"fmt"
)

type FieldTypeName string

const (
	FieldType_String        FieldTypeName = "string"
	FieldType_Id            FieldTypeName = "id"
	FieldType_Ref           FieldTypeName = "ref"
	FieldType_Array         FieldTypeName = "array"
	FieldType_Datetime      FieldTypeName = "datetime"
	FieldType_Date          FieldTypeName = "date"
	FieldType_Time          FieldTypeName = "time"
	FieldType_Int           FieldTypeName = "int"
	FieldType_Range         FieldTypeName = "range"
	FieldType_Bool          FieldTypeName = "bool"
	FieldType_Text          FieldTypeName = "text"
	FieldType_Address       FieldTypeName = "address"
	FieldType_Float         FieldTypeName = "float"
	FieldType_Password      FieldTypeName = "password"
	FieldType_File          FieldTypeName = "file"
	FieldType_Enum          FieldTypeName = "enum"
	FieldType_Color         FieldTypeName = "color"
	FieldType_Autotimestamp FieldTypeName = "autotimestamp"
	FieldType_Timestamp     FieldTypeName = "timestamp"
	FieldType_Sqltimestamp  FieldTypeName = "sqltimestamp"
	FieldType_Bitswitch     FieldTypeName = "bitswitch"
	FieldType_Keyval        FieldTypeName = "keyval"
	FieldType_Blobject      FieldTypeName = "blobject"
	FieldType_Gob           FieldTypeName = "gob"
	FieldType_Refid         FieldTypeName = "refid"
	//FieldType_Patientcard FieldTypeName = "patientcard"
)

type FieldType interface {
	FromDb(interface{}) (interface{}, error)
	ToDb(interface{}) (string, error)
	GetScanReciever() interface{}
	IsSearchable() bool
	GetMysqlDef() string
}

func FieldByType(fieldTypeName FieldTypeName) (FieldType, error) {

	switch fieldTypeName {
	case FieldType_String:
		return &FieldString{}, nil
	case FieldType_Id:
		return &FieldId{}, nil
	case FieldType_Ref:
		return &FieldRef{}, nil
	case FieldType_Array:
		return &FieldString{}, nil
	case FieldType_Datetime:
		return &FieldDateTime{}, nil
	case FieldType_Date:
		return &FieldDate{}, nil
	case FieldType_Time:
		return &FieldTime{}, nil
	case FieldType_Int:
		return &FieldInt{}, nil
	case FieldType_Range:
		return &FieldInt{}, nil
	case FieldType_Bool:
		return &FieldBool{}, nil
	case FieldType_Text:
		return &FieldText{}, nil
	case FieldType_Address:
		return &FieldText{}, nil
	case FieldType_Float:
		return &FieldFloat{}, nil
	case FieldType_Password:
		return &FieldPassword{}, nil
	case FieldType_File:
		return &FieldFile{}, nil
	case FieldType_Enum:
		return &FieldEnum{}, nil
	case FieldType_Color:
		return &FieldString{}, nil
	case FieldType_Autotimestamp:
		return &FieldInt{}, nil
	case FieldType_Timestamp:
		return &FieldTimestamp{}, nil
	case FieldType_Sqltimestamp:
		return &FieldTimestamp{}, nil
	case FieldType_Bitswitch:
		return &FieldInt{}, nil
	case FieldType_Keyval:
		return &FieldKeyVal{}, nil
	case FieldType_Blobject:
		return &FieldBlobject{}, nil
	case FieldType_Gob:
		return &FieldGob{}, nil
	case FieldType_Refid:
		return &FieldRefID{}, nil
	default:
		return nil, errors.New(fmt.Sprintf("Invalid Field Type '%s'", fieldTypeName))
	}
}
