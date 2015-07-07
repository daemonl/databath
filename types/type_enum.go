package types

import (
	"encoding/xml"
	"fmt"
)

// string
type FieldEnum struct {
	Length  int
	Choices map[string]string
}

func (f *FieldEnum) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	xmlRepresentation := &struct {
		Length  int `xml:"length,attr"`
		Choices []struct {
			Key   string `xml:"key,attr"`
			Value string `xml:",innerxml"`
		} `xml:"choice"`
	}{}
	err := d.DecodeElement(xmlRepresentation, &start)
	if err != nil {
		return err
	}
	f.Length = xmlRepresentation.Length
	f.Choices = map[string]string{}
	for _, choice := range xmlRepresentation.Choices {
		f.Choices[choice.Key] = choice.Value
		if f.Length < len(choice.Key) {
			f.Length = len(choice.Key)
		}
	}
	return nil
}

func (f *FieldEnum) GetMysqlDef() string {
	return fmt.Sprintf("VARCHAR(%d) NULL", f.Length)
}

func (f *FieldEnum) IsSearchable() bool { return true }

func (f *FieldEnum) FromDb(stored interface{}) (interface{}, error) {
	// String -> String

	storedStringPointer, ok := stored.(*string)
	if !ok {
		return nil, makeConversionError("string", stored)
	}

	if storedStringPointer == nil {
		return nil, nil
	} else {
		return UnescapeString(*storedStringPointer), nil
	}
}

func (f *FieldEnum) ToDb(input interface{}) (string, error) {
	// String -> String
	inputString, ok := input.(string)
	if !ok {
		return "", MakeToDbUserErrorFromString(fmt.Sprintf("Converting string to DB, Value Must be a string, got '%v'", input))
	}
	return EscapeString(inputString), nil
}
func (f *FieldEnum) GetScanReciever() interface{} {
	var s string
	var sp *string = &s
	return &sp
}
