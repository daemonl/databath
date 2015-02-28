package types

//////////////
// GOB //
//////////////
type FieldGob struct{}

func (f *FieldGob) GetMysqlDef() string { return "BLOB NULL" }

func (f *FieldGob) IsSearchable() bool { return true }

func (f *FieldGob) Init(raw map[string]interface{}) error { return nil }

func (f *FieldGob) FromDb(stored interface{}) (interface{}, error) {
	// String -> String
	storedString, ok := stored.(*[]byte)
	if !ok {
		return nil, MakeFromDbErrorFromString("Incorrect Type in DB (expected []byte)")
	}
	if storedString == nil {
		return nil, nil
	}
	return storedString, nil
}

func (f *FieldGob) ToDb(input interface{}) (string, error) {
	// String -> String
	inputString, ok := input.([]byte)
	if !ok {
		return "", MakeToDbUserErrorFromString("Must be a []byte")
	}
	return string(inputString), nil
}
func (f *FieldGob) GetScanReciever() interface{} {
	var v []byte
	var vp *[]byte = &v
	return &vp
}
