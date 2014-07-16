package types

import ()

//////////////
// BLOBJECT //
//////////////
type FieldBlobject struct{}

func (f *FieldBlobject) GetMysqlDef() string { return "TEXT NULL" }

func (f *FieldBlobject) IsSearchable() bool { return true }

func (f *FieldBlobject) Init(raw map[string]interface{}) error { return nil }

func (f *FieldBlobject) FromDb(stored interface{}) (interface{}, error) {
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
func (f *FieldBlobject) ToDb(input interface{}) (string, error) {
	// String -> String
	inputString, ok := input.(string)
	if !ok {
		return "", MakeToDbUserErrorFromString("Must be a string")
	}
	return EscapeString(inputString), nil
}
func (f *FieldBlobject) GetScanReciever() interface{} {
	var v string
	var vp *string = &v
	return &vp
}
