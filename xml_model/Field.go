package xml_model

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/daemonl/databath"
	"github.com/daemonl/databath/types"
)

type Field struct {
	Name string
	Type types.FieldTypeName
	Impl types.FieldType
	Raw  map[string]interface{}
}

func (raw *Field) ToDatabath() (*databath.Field, error) {
	field := &databath.Field{}
	field.Raw = raw.Raw
	field.Path = raw.Name
	field.FieldType = raw.Impl
	return field, nil
}

func (f *Field) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	f.Raw = map[string]interface{}{}
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "type":
			typeString := attr.Value
			typeString = strings.Replace(typeString, "_", "", -1)
			typeString = strings.ToLower(typeString)
			f.Type = types.FieldTypeName(attr.Value)
		case "name":
			f.Name = strings.ToLower(attr.Value)
			//default:
			//f.Raw[attr.Name] = attr.Value
		}
	}
	if len(f.Name) < 1 {
		return fmt.Errorf("No 'name' attribute on field element")
	}
	if len(f.Type) < 1 {
		return fmt.Errorf("No 'type' attribute on field element %s", f.Name)
	}

	fieldType, err := types.FieldByType(f.Type)
	if err != nil {
		return err
	}
	f.Impl = fieldType
	return d.DecodeElement(f.Impl, &start)
}
