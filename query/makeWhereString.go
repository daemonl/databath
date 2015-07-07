package query

import (
	"log"
	"strconv"
	"strings"
)

func (q *Query) makeWhereString(conditions *QueryConditions) (whereString string, whereParameters []interface{}, havingString string, havingParameters []interface{}, returnErr error) {
	log.Println("Begin makeWhereString")

	whereString = ""
	havingString = ""
	whereParameters = make([]interface{}, 0, 0)
	havingParameters = make([]interface{}, 0, 0)

	if conditions.where == nil {
		conditions.where = make([]QueryCondition, 0, 0)
		log.Println("Add empty conditions.where")
	}

	if conditions.pk != nil {
		log.Println("Add PK condition")
		pkCondition := QueryConditionWhere{
			Field: "id",
			Cmp:   "=",
			Val:   *conditions.pk,
		}
		conditions.where = append(conditions.where, &pkCondition)
	}

	if conditions.filter != nil {
		for fieldName, value := range *conditions.filter {
			fieldNames := strings.Split(fieldName, ",")
			qcArray := []QueryCondition{}
			for _, field := range fieldNames {
				qc := &QueryConditionWhere{
					Field: strings.TrimSpace(field),
					Cmp:   "=",
					Val:   value,
				}
				qcArray = append(qcArray, qc)
			}
			joined, joinedParameters, _, _, err := q.JoinConditionsWith(qcArray, " OR ")
			if err != nil {
				returnErr = err
				return //BAD
			}
			if len(joined) > 0 {
				strCondition := QueryConditionString{Str: joined, Parameters: joinedParameters}
				conditions.where = append(conditions.where, &strCondition)
			}
		}
	}

	// Search for things
	if conditions.search != nil {

		for field, term := range conditions.search {

			parts := re_notAlphaNumeric.Split(strings.TrimSpace(term), -1)
			blobjectPairings := re_blobjectNotAlphaNumeric.Split(strings.TrimSpace(term), -1)

			if field == "*" {
				// Search for a number - Make it an ID:
				if re_numeric.MatchString(term) {
					number, _ := strconv.ParseUint(term, 10, 32)
					filterCondition := QueryConditionWhere{
						Field: "id",
						Cmp:   "=",
						Val:   number,
					}
					conditions.where = append(conditions.where, &filterCondition)
					continue
				}

				var usePrefixSearch bool
				for pString, searchPrefix := range q.collection.SearchPrefixes {
					if strings.HasPrefix(term, pString) {
						termWithoutPrefix := term[len(pString):]
						if re_numeric.MatchString(termWithoutPrefix) {
							usePrefixSearch = true
							number, _ := strconv.ParseUint(termWithoutPrefix, 10, 32)
							filterCondition := QueryConditionWhere{
								Field: searchPrefix.FieldName,
								Cmp:   "LIKE",
								Val:   number,
							}
							conditions.where = append(conditions.where, &filterCondition)
							break
						}
					}
				}

				if usePrefixSearch {
					continue
				}

				for _, part := range parts {
					partGroup := make([]QueryCondition, 0, 0)
					for _, mappedField := range q.map_field {
						condition := mappedField.ConstructQuery(part)
						if condition != nil {
							partGroup = append(partGroup, condition)
						}
					}
					j1, jp1, _, _, err := q.JoinConditionsWith(partGroup, " OR ")
					if err != nil {
						returnErr = err
						return //BAD
					}
					strCondition := QueryConditionString{Str: j1, Parameters: jp1}
					conditions.where = append(conditions.where, &strCondition)
				}
			} else {
				fieldNames := strings.Split(field, ",")
				partGroup := make([]QueryCondition, 0, len(parts)*len(fieldNames))
				for _, p := range blobjectPairings {
					if p != "" {
						for _, field := range fieldNames {
							mappedField, err := q.getMappedFieldByFieldName(field)
							if err != nil {
								returnErr = err
								return //BAD
							}
							qc := mappedField.ConstructQuery(p)
							if qc != nil {
								partGroup = append(partGroup, qc)
							}
						}
					}
				}
				joined, joinedParameters, _, _, err := q.JoinConditionsWith(partGroup, " AND ")
				if err != nil {
					returnErr = err
					return //BAD
				}
				if len(joined) > 0 {
					strCondition := QueryConditionString{Str: joined, Parameters: joinedParameters}
					conditions.where = append(conditions.where, &strCondition)
				}
			}
		}
	}

	whereString, whereParameters, havingString, havingParameters, err := q.JoinConditionsWith(conditions.where, " AND ")
	if err != nil {
		returnErr = err
		return //BAD
	}
	if len(whereString) > 1 {
		whereString = "WHERE " + whereString
	}
	if len(havingString) > 1 {
		havingString = "HAVING " + havingString
	}

	return //GOOD
}
