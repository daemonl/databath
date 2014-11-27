package databath

import (
	"fmt"
	"reflect"
	"strings"
)

type QueryConditionWhere struct {
	Field string      `json:"field"`
	Cmp   string      `json:"cmp"`
	Val   interface{} `json:"val"`
}

func (qc *QueryConditionWhere) GetConditionString(q *Query) (queryString string, parameters []interface{}, isAggregate bool, returnErr error) {

	queryString = ""
	parameters = make([]interface{}, 0, 0)
	isAggregate = false
	returnErr = nil
	queryUsingName := ""

	field, ok := q.map_field[qc.Field]
	if !ok {
		q.Dump()
		returnErr = UserErrorF("Cannot query on non mapped field '%s'.", qc.Field)
		return //BAD
	}

	switch f := field.fieldSetFieldDef.(type) {
	case *FieldSetFieldDefNormal:
		isAggregate = false
		queryUsingName = fmt.Sprintf("`%s`.`%s`", field.table.alias, field.fieldNameInTable)
	case *FieldSetFieldDefRaw:
		if f.SearchOn != nil {
			isAggregate = false
			queryUsingName = field.table.alias + "." + *f.SearchOn
		} else {
			isAggregate = true
			queryUsingName = field.alias
		}
	default:
		isAggregate = true
		queryUsingName = field.alias
	}

	valString, ok := qc.Val.(string)
	if ok && strings.HasPrefix(valString, "#") {
		paramName := valString[1:]
		qc.Val = q.context.GetValueFor(paramName)
	}
	if qc.Cmp == "IN" {

		switch reflect.TypeOf(qc.Val).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(qc.Val)
			length := s.Len()
			escapedSlice := make([]string, length, length)
			for i := 0; i < length; i++ {
				dbVal, err := field.field.ToDb(s.Index(i).Interface(), q.context)
				if err != nil {
					return
				}
				escapedSlice[i] = "?"
				parameters = append(parameters, dbVal)
			}
			queryString = fmt.Sprintf("%s IN (%s)", queryUsingName, strings.Join(escapedSlice, ", "))
			return //GOOD

		default:
			fmt.Printf("TYPE for IN: %v\n", qc.Val)
			returnErr = UserErrorF("IN conditions require that val is an array")
			return //BAD
		}

	} else if qc.Cmp == "=" || qc.Cmp == "!=" || qc.Cmp == "<=" || qc.Cmp == ">=" || qc.Cmp == "<" || qc.Cmp == ">" {
		dbVal, err := field.field.ToDb(qc.Val, q.context)
		if err != nil {
			returnErr = UserErrorF("%T.ToDb Error: %s", field.field, err.Error())
			return
		}
		parameters = append(parameters, dbVal)
		queryString = fmt.Sprintf("%s %s ?", queryUsingName, qc.Cmp)
		return //GOOD
	} else if qc.Cmp == "LIKE" {
		dbVal, err := field.field.ToDb(qc.Val, q.context)
		if err != nil {
			returnErr = err
			return //BAD
		}
		dbVal = "%" + dbVal + "%"
		parameters = append(parameters, dbVal)
		queryString = fmt.Sprintf("%s LIKE ?", queryUsingName)
		return //GOOD
	} else if qc.Cmp == "INJSON" {
		dbVal, err := field.field.ToDb(qc.Val, q.context)
		blobjectPairing := strings.Split(dbVal, ":")
		fmt.Println("-------------------------------")
		fmt.Println(blobjectPairing)
		if err != nil {
			returnErr = err
			return //BAD
		}
		//This regexp returns all rows with values including a word starting with "dbVal".
		// NB: The mysql regexp implementation is weird - \b is replaced with [[:<:]], and
		//     lookahead/lookbehind aren't supported.
		fmt.Printf("length %v\n", len(blobjectPairing))
		fmt.Printf("pairing %v\n", blobjectPairing)
		if len(blobjectPairing) == 2 {
			parameter := blobjectPairing[0] + "\":\"" + blobjectPairing[1] + "[^\"]*((\",)|(\"\\}))"
			parameters = append(parameters, parameter)
			queryString = fmt.Sprintf("%s REGEXP ?", queryUsingName)
			return //GOOD
		}
		returnErr = UserErrorF("invalid blobject search")
		return //BAD
	} else if qc.Cmp == "IS NULL" || qc.Cmp == "IS NOT NULL" {
		queryString = fmt.Sprintf("%s %s", queryUsingName, qc.Cmp)
		return //GOOD
	} else {
		returnErr = UserErrorF("Compare method not allowed")
		return //BAD
	}

}
