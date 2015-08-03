package query_builder

import "testing"

type testCollection struct {
	name   string
	fields []QueryField
}

func (tc *testCollection) TableName() string {
	return tc.name
}
func (tc *testCollection) GetField(name string) (QueryField, bool) {
	// BAD the func signature requires the QB library.
	for _, f := range tc.fields {
		if f.FieldName() == name {
			return f, true
		}
	}
	return nil, false
}

type testField struct {
	name string
}

func (tf *testField) FieldName() string {
	return tf.name
}

type testRefField struct {
	*testField
	refCollection QueryCollection
}

func (trf *testRefField) RefCollection() QueryCollection {
	return trf.refCollection

}

func TestQueryBuilder(t *testing.T) {

	cOrganisation := &testCollection{
		name: "organisation",
		fields: []QueryField{
			&testField{name: "name"},
			&testField{name: "phone"},
		},
	}
	cPerson := &testCollection{
		name: "person",
		fields: []QueryField{
			&testField{name: "name"},
			&testField{name: "email"},
			&testRefField{
				testField: &testField{
					name: "organisation",
				},
				refCollection: cOrganisation,
			},
		},
	}

	qb := NewQueryBuilder()
	root := qb.From(cPerson)
	organisation := root.LeftJoin(cPerson.fields[2].(RefField))

	root.AddName("email")
	organisation.AddName("name")
	organisation.AddName("phone")

	qr, _ := qb.GetQuery()
	q := qr.(*query)

	if q.sqlCols != "SELECT T0.email AS F0, T1.name AS F1, T1.phone AS F2 FROM person T0 LEFT JOIN organisation T1 ON T0.organisation = T1.id  GROUP BY t0.id " {
		t.Log(q.sqlCols)
		t.Fail()
	}
}
