package xml_model

import (
	"fmt"

	"github.com/daemonl/databath/model"
)

type Collection struct {
	Name           string           `xml:"name,attr"`
	Fields         []Field          `xml:"field"`
	Views          []View           `xml:"view"`
	CustomFields   []CustomField    `xml:"custom"`
	SearchPrefixes []SearchPrefix   `xml:"searchPrefixe"`
	Masks          []CollectionMask `xml:"mask"`
	//	ViewQuery      *string          `json:"viewQuery,omitempty"`
}

func (raw *Collection) ToDatabath() (*model.Collection, error) {
	dbc := &databath.Collection{}
	dbc.TableName = raw.Name
	dbc.Fields = map[string]*databath.Field{}
	for _, rawField := range raw.Fields {
		field, err := rawField.ToDatabath()
		if err != nil {
			return nil, err
		}
		field.Collection = dbc
		dbc.Fields[rawField.Name] = field
	}

	_, ok := dbc.Fields["id"]
	if !ok {
		return nil, fmt.Errorf("Error parsing collection %s, no id field", raw.Name)
	}

	return dbc, nil
}
