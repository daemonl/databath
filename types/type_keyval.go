package types

import ()

////////////
// KEYVAL //
////////////
type FieldKeyVal struct{}

func (f *FieldKeyVal) GetMysqlDef() string { return "TEXT NULL" }

func (f *FieldKeyVal) IsSearchable() bool { return true }

func (f *FieldKeyVal) Init(raw map[string]interface{}) error { return nil }

func (f *FieldKeyVal) FromDb(stored interface{}) (interface{}, error) {
	// String -> String
	storedString, ok := stored.(*string)
	if !ok {
		return nil, MakeFromDbErrorFromString("Incorrect Type in DB (expected string)")
	}
	if storedString == nil {
		return nil, nil
	}
	return UnescapeString(*storedString), nil
}
func (f *FieldKeyVal) ToDb(input interface{}) (string, error) {
	// String -> String
	inputString, ok := input.(string)
	if !ok {
		return "", MakeToDbUserErrorFromString("Must be a string")
	}
	return EscapeString(inputString), nil
}
func (f *FieldKeyVal) GetScanReciever() interface{} {
	var v string
	var vp *string = &v
	return &vp
}
