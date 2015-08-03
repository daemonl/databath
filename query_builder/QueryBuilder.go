package query_builder

import (
	"fmt"
	"strings"
	"sync"
)

type QueryBuilder struct {
	from          *table
	columns       []*column
	tables        []*table
	joins         []string
	nextColumn    int
	nextTable     int
	countLock     sync.Mutex
	baseCondition QueryCondition
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		columns:   []*column{},
		tables:    []*table{},
		joins:     []string{},
		countLock: sync.Mutex{},
		baseCondition: &WhereGroup{
			joiner: "AND",
			parts:  []QueryCondition{},
		},
	}
}

func (qb *QueryBuilder) NextFieldAlias() string {
	qb.countLock.Lock()
	alias := fmt.Sprintf("F%d", qb.nextColumn)
	qb.nextColumn += 1
	qb.countLock.Unlock()
	return alias
}

func (qb *QueryBuilder) NextTableAlias() string {
	qb.countLock.Lock()
	alias := fmt.Sprintf("T%d", qb.nextTable)
	qb.nextTable += 1
	qb.countLock.Unlock()
	return alias
}

type ErrFieldNotFound struct {
	Table string
	Field string
}

func (e *ErrFieldNotFound) Error() string {
	return fmt.Sprintf("Field %s.%s not found", e.Table, e.Field)
}

type QueryCondition interface {
	FleshOut() (string, []interface{})
}
type WhereGroup struct {
	parts  []QueryCondition
	joiner string
}

func (wg *WhereGroup) FleshOut() (string, []interface{}) {
	if len(wg.parts) < 1 {
		return "", []interface{}{}
	}
	params := []interface{}{}
	parts := make([]string, len(wg.parts), len(wg.parts))
	for i, p := range wg.parts {
		newPart, newParams := p.FleshOut()
		params = append(params, newParams...)
		parts[i] = newPart
	}
	return "(" + strings.Join(parts, wg.joiner) + ")", params
}

type QueryCollection interface {
	TableName() string
	GetField(string) (QueryField, bool)
}

type table struct {
	QueryCollection
	builder *QueryBuilder
	alias   string
}

type Table interface {
	LeftJoin(RefField) Table
	Add(QueryField) error
	AddName(string) error
	//LeftJoinName(string) error
}

func (t *table) LeftJoin(field RefField) Table {
	refCollection := field.RefCollection()
	joined := t.builder.includeCollection(refCollection)
	joinstring := fmt.Sprintf(
		"LEFT JOIN %s %s ON %s.%s = %s.id",
		joined.QueryCollection.TableName(),
		joined.alias,
		t.alias,
		field.FieldName(),
		joined.alias,
	)
	t.builder.joins = append(t.builder.joins, joinstring)
	return joined
}

func (t *table) Add(field QueryField) error {
	alias := t.builder.NextFieldAlias()
	t.builder.columns = append(t.builder.columns, &column{
		QueryField: field,
		alias:      alias,
		table:      t,
	})
	return nil
}

func (t *table) AddName(fieldName string) error {
	field, ok := t.QueryCollection.GetField(fieldName)
	if !ok {
		return &ErrFieldNotFound{
			Table: t.QueryCollection.TableName(),
			Field: fieldName,
		}
	}
	return t.Add(field)
}

type RefField interface {
	FieldName() string
	RefCollection() QueryCollection
}

func (qb *QueryBuilder) From(collection QueryCollection) Table {
	qb.from = qb.includeCollection(collection)
	return qb.from
}

type column struct {
	QueryField
	alias string
	table *table
}

type QueryField interface {
	FieldName() string
}

func (qb *QueryBuilder) includeCollection(collection QueryCollection) *table {
	qb.countLock.Lock()
	alias := fmt.Sprintf("T%d", qb.nextTable)
	qb.nextTable += 1
	qb.countLock.Unlock()
	table := &table{
		QueryCollection: collection,
		alias:           alias,
		builder:         qb,
	}
	qb.tables = append(qb.tables, table)
	return table
}

func (qb *QueryBuilder) selectFields() []string {
	parts := make([]string, len(qb.columns), len(qb.columns))
	for i, col := range qb.columns {
		parts[i] = fmt.Sprintf("%s.%s AS %s", col.table.alias, col.QueryField.FieldName(), col.alias)
	}
	return parts
}

type query struct {
	sqlCols  string
	sqlCount string
	params   []interface{}
}
type Query interface{}

func (qb *QueryBuilder) GetQuery() (Query, error) {
	q := &query{}

	joinString := strings.Join(qb.joins, " ")
	whereString, whereParams := qb.baseCondition.FleshOut()

	pageString := ""
	q.params = whereParams

	q.sqlCols = fmt.Sprintf(`SELECT %s FROM %s T0 %s %s GROUP BY t0.id %s`,
		strings.Join(qb.selectFields(), ", "),
		qb.from.TableName(),
		joinString,
		whereString,
		pageString)

	q.sqlCount = fmt.Sprintf(`SELECT COUNT(t0.id) FROM %s t0 %s %s GROUP BY t0.id`,
		qb.from.TableName(),
		joinString,
		whereString)

	return q, nil
}
