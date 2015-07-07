package types

//////////
// bool //
//////////
type FieldBool struct {
	Optional bool `xml:"optional,attr" json:"optional"`
}

func (f *FieldBool) GetMysqlDef() string {

	def := "TINYINT(1)"

	if !f.Optional {
		def = def + " NOT NULL"
	} else {
		def = def + " NULL"
	}

	return def

}

func (f *FieldBool) IsSearchable() bool { return false }

func (f *FieldBool) FromDb(stored interface{}) (interface{}, error) {
	storedBool, ok := stored.(*bool)
	if !ok {
		return nil, MakeFromDbErrorFromString("Incorrect Type in DB (expecting bool)")
	}
	if storedBool == nil {
		return nil, nil
	}
	return *storedBool, nil
}

func (f *FieldBool) ToDb(input interface{}) (string, error) {
	switch input := input.(type) {
	case bool:
		if input {
			return "1", nil
		} else {
			return "0", nil
		}

	default:
		return "", makeConversionError("bool", input)
	}
}

func (f *FieldBool) GetScanReciever() interface{} {
	var v bool
	var vp *bool = &v
	return &vp
}
