package types

import ()

type FieldRefID struct{
Collection string
}

func (f *FieldRefID) GetMysqlDef() string { return "INT(11) UNSIGNED NOT NULL" }

func (f *FieldRefID) IsSearchable() bool { return false }

func (f *FieldRefID) Init(raw map[string]interface{}) error { 

	collection, ok := raw["collection"]
	if !ok{
		return ModelDefinitionError{"RefID Field nas no key 'collection'", ""}
	}
	collectionString := collection.(string)
	f.Collection = collectionString

	return nil
}

func (f *FieldRefID) FromDb(stored interface{}) (interface{}, error) {
	// uInt64 -> uInt64
	storedInt, ok := stored.(*uint64)
	if !ok {
		return nil, MakeFromDbErrorFromString("Incorrect Type in DB (expected uint64)")
	}
	if storedInt == nil {
		return nil, nil
	}
	return *storedInt, nil
}
func (f *FieldRefID) ToDb(input interface{}) (string, error) {
	// uInt64 -> uInt64
	return unsignedIntToDb(input)
}
func (f *FieldRefID) GetScanReciever() interface{} {
	var v uint64
	var vp *uint64 = &v
	return &vp
}
