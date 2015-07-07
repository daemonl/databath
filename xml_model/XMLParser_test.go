package xml_model

import (
	"encoding/xml"
	"testing"

	"github.com/daemonl/databath/types"
)

func TestField(t *testing.T) {
	field := &Field{}
	err := xml.Unmarshal([]byte(`
	<field name="bob" type="string"/>
	`), field)
	if err != nil {
		t.Error(err)
		return
	}
	if field.Name != "bob" {
		t.Fail()
		return
	}
	if field.Type != "string" {
		t.Fail()
		return
	}

	field = &Field{}
	err = xml.Unmarshal([]byte(`
	<field name="bob1" type="bool" optional="true"/>
	`), field)
	if err != nil {
		t.Error(err)
		return
	}
	if field.Name != "bob1" {
		t.Fail()
	}
	boolField, ok := field.Impl.(*types.FieldBool)
	if !ok {
		t.Fail()
		return
	}
	if boolField.Optional == false {
		t.Fail()
		return
	}

	field = &Field{}
	err = xml.Unmarshal([]byte(`
	<field name="bob3" type="enum">
		<choice key='a'>First Option</choice>
		<choice key='b'>Second Option</choice>
	</field>
	`), field)
	if err != nil {
		t.Error(err)
		return
	}

	enumField, ok := field.Impl.(*types.FieldEnum)
	if !ok {
		t.Fail()
		return
	}
	t.Log(enumField)

	if enumField.Choices["a"] != "First Option" {
		t.Fail()
		return
	}
	if enumField.Choices["b"] != "Second Option" {
		t.Fail()
		return
	}
	if enumField.Length != 1 {
		t.Fail()
		return
	}

	field = &Field{}
	err = xml.Unmarshal([]byte(`
	 <field name="primary_contact" type="ref" collection="person">
      <limit path="company.id">id</limit>
    </field>
	`), field)
	if err != nil {
		t.Error(err)
		return
	}
	refField, ok := field.Impl.(*types.FieldRef)
	t.Log(refField)
	if !ok {
		t.Fail()
		return
	}

	if refField.Collection != "person" {
		t.Fail()
		return
	}

	if refField.Limit["company.id"] != "id" {
		t.Fail()
		return
	}

}
