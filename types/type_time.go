package types

import ()

type FieldTime struct{}

func (f *FieldTime) GetMysqlDef() string { return "TIME NULL" }

func (f *FieldTime) IsSearchable() bool { return false }

func (f *FieldTime) FromDb(stored interface{}) (interface{}, error) {
	//

	storedString, ok := stored.(*string)
	if !ok {
		return nil, makeConversionError("time", stored)
	}

	if storedString == nil {
		return nil, nil
	}
	return *storedString, nil

}

func (f *FieldTime) ToDb(input interface{}) (string, error) {

	str, ok := input.(string)
	if !ok {
		return "", MakeToDbUserErrorFromString("Must be a string")
	}

	return str, nil
}

func (f *FieldTime) GetScanReciever() interface{} {
	var v string
	var vp *string = &v
	return &vp
}
