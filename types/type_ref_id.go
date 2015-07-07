package types

import ()

type FieldRefID struct {
	Collection string `xml:"collection,attr"`
}

func (f *FieldRefID) GetMysqlDef() string { return "INT(11) UNSIGNED NOT NULL" }

func (f *FieldRefID) IsSearchable() bool { return false }

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
