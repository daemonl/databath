package types

import "encoding/xml"

type FieldRef struct {
	Collection string
	OnDelete   int
	Limit      map[string]interface{}
}

const (
	RefOnDeletePrevent = iota
	RefOnDeleteCascade
	RefOnDeleteNull
)

func (f *FieldRef) GetMysqlDef() string { return "INT(11) UNSIGNED NULL" }

func (f *FieldRef) IsSearchable() bool { return false }

func (f *FieldRef) GetRefCollectionName() string {
	return f.Collection
}
func (f *FieldRef) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	xmlRepresentation := &struct {
		Collection string `xml:"collection,attr"`
		OnDelete   string `xml:"on_delete",attr"`
		Limits     []struct {
			Path  string `xml:"path,attr"`
			Value string `xml:",innerxml"`
		} `xml:"limit"`
	}{}
	err := d.DecodeElement(xmlRepresentation, &start)
	if err != nil {
		return err
	}
	f.Collection = xmlRepresentation.Collection
	switch xmlRepresentation.OnDelete {
	case "CASCADE":
		f.OnDelete = RefOnDeleteCascade
	case "NULL":
		f.OnDelete = RefOnDeleteNull
	default:
		f.OnDelete = RefOnDeletePrevent
	}
	f.Limit = map[string]interface{}{}
	for _, limit := range xmlRepresentation.Limits {
		f.Limit[limit.Path], err = InferTypeFromString(limit.Value)
		if err != nil {
			return nil
		}
	}
	return nil
}

func (f *FieldRef) FromDb(stored interface{}) (interface{}, error) {
	// uInt64 -> Iunt64
	storedInt, ok := stored.(*uint64)
	if !ok {
		return nil, MakeFromDbErrorFromString("Incorrect Type in DB (expected uint64)")
	}
	if storedInt == nil {
		return nil, nil
	}
	return *storedInt, nil
}

func (f *FieldRef) ToDb(input interface{}) (string, error) {
	return unsignedIntToDb(input)
}

func (f *FieldRef) GetScanReciever() interface{} {
	var v uint64
	var vp *uint64 = &v
	return &vp
}
