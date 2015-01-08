package types

import ()

//////////////
// BLOBJECT //
//////////////
type FieldPatientCard struct{}

func (f *FieldPatientCard) GetMysqlDef() string { return "TEXT NULL" }

func (f *FieldPatientCard) IsSearchable() bool { return true }

func (f *FieldPatientCard) Init(raw map[string]interface{}) error { return nil }

func (f *FieldPatientCard) FromDb(stored interface{}) (interface{}, error) {
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
func (f *FieldPatientCard) ToDb(input interface{}) (string, error) {
	// String -> String
	inputString, ok := input.(string)
	if !ok {
		return "", MakeToDbUserErrorFromString("Must be a string")
	}
	return EscapeString(inputString), nil
}
func (f *FieldPatientCard) GetScanReciever() interface{} {
	var v string
	var vp *string = &v
	return &vp
}
