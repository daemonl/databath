package databath

import "database/sql"

type QueryBuilder interface {
	Select(...QueryColumn) QueryBuilder
	From(QueryCollection) QueryBuilder
	Join(QueryJoin) QueryBuilder
	Where(QueryCondition) QueryBuilder
	GroupBy(QueryField) QueryBuilder
	OrderBy(QueryField) QueryBuilder
	Limit(int) QueryBuilder
	Offset(int) QueryBuilder

	GetQuery() Query
}

type Query interface {
	Run(db *sql.DB)
}

// QueryField represents a field in a table
type QueryField interface {
}

// QueryColumn 'extends' query field, adding display
// info for the end user
type QueryColumn interface {
	QueryField
	SetAlias(string)
	BuildScanPointer() interface{}
	FormatScanPointer(interface{}) interface{}
	GetTitle() string
}

type QueryCollection interface{}

// QueryJoin is a join to add... to the query
type QueryJoin interface {
	GetJoinString() string
	QueryCollection
}

// QueryCondition resolves to part of a WHERE clause.
// all query conditions are ANDed together.
type QueryCondition interface {
	GetWhereString() string
}
