package sync

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/daemonl/databath"
	"github.com/daemonl/databath/types"
)

func doErr(err error) {
	if err != nil {
		panic(err)
	}
}

type TableStatus struct {
	Name            string  `sql:"Name"`
	Engine          string  `sql:"Engine"`
	Version         *uint64 `sql:"Version"`
	Row_format      *string `sql:"Row_format"`
	Rows            *uint64 `sql:"Rows"`
	Avg_row_length  *uint64 `sql:"Avg_row_length"`
	Data_length     *uint64 `sql:"Data_length"`
	Max_data_length *uint64 `sql:"Max_data_length"`
	Index_length    *uint64 `sql:"Index_length"`
	Data_free       *uint64 `sql:"Data_free"`
	Auto_increment  *uint64 `sql:"Auto_increment"`
	Create_time     *string `sql:"Create_time"`
	Update_time     *string `sql:"Update_time"`
	Check_time      *string `sql:"Check_time"`
	Collation       *string `sql:"Collation"`
	Checksum        *string `sql:"Checksum"`
	Create_options  *string `sql:"Create_options"`
	Comment         *string `sql:"Comment"`
}

type Column struct {
	Field   string  `sql:"Field"`
	Type    string  `sql:"Type"`
	Null    string  `sql:"Null"`
	Key     *string `sql:"Key"`
	Default *string `sql:"Default"`
	Extra   *string `sql:"Extra"`
}

type Index struct {
	ConstraintName       string  `sql:"CONSTRAINT_NAME"`
	TableName            *string `sql:"TABLE_NAME"`
	ConstraintType       *string `sql:"CONSTRAINT_TYPE"`
	ColumnName           *string `sql:"COLUMN_NAME"`
	ReferencedTableName  *string `sql:"REFERENCED_TABLE_NAME"`
	ReferencedColumnName *string `sql:"REFERENCED_COLUMN_NAME"`
	Used                 bool
}

type SyncError struct {
	Message string
}

func (e *SyncError) Error() string {
	return e.Message
}

var execString string = ""
var unused []string = []string{}

var reMidWhitespace *regexp.Regexp = regexp.MustCompile(`[\n\t\ ]+`)
var reLeadingWhitespace *regexp.Regexp = regexp.MustCompile(`^[\n\t\ ]+`)
var reTrailingWhitespace *regexp.Regexp = regexp.MustCompile(`[\n\t\ ]+$`)

var reCheckLength *regexp.Regexp = regexp.MustCompile(`^VARCHAR\(([0-9]+)\)`)

func (c *Column) GetString() string {
	built := c.Type
	if c.Null == "NO" {
		built += " NOT NULL"
	} else {
		built += " NULL"
	}
	if c.Extra != nil {
		built += " " + *c.Extra
	}
	built = strings.TrimSpace(built)
	return strings.ToUpper(built)
}

func ScanToStruct(res *sql.Rows, obj interface{}, tag string) error {
	if len(tag) == 0 {
		tag = "sql"
	}
	rv := reflect.ValueOf(obj)
	rt := reflect.TypeOf(obj)

	if reflect.Indirect(rv).Kind().String() != "struct" {
		panic("KIND NOT STRUCT" + rv.Kind().String())
	}

	valueElm := rv.Elem()

	maxElements := rt.Elem().NumField()
	//scanVals := make([]interface{}, maxElements, maxElements)
	cols, err := res.Columns()
	if err != nil {
		return err
	}

	scanVals := map[string]interface{}{}

	for i := 0; i < maxElements; i++ {
		interf := valueElm.Field(i).Addr().Interface()
		sqlTag := valueElm.Type().Field(i).Tag.Get(tag)
		if len(sqlTag) < 1 || sqlTag == "-" {
			continue
		}
		scanVals[sqlTag] = interf
	}
	orderedScanVals := make([]interface{}, len(cols), len(cols))
	for i, colName := range cols {
		interfPtr, ok := scanVals[colName]
		if !ok {
			return &SyncError{fmt.Sprintf("SQL Column '%s' had no matching struct tag", colName)}
		}
		orderedScanVals[i] = interfPtr
	}
	err = res.Scan(orderedScanVals...)

	if err != nil {
		return err
	}
	return nil
}

func MustExecF(now bool, db *sql.DB, format string, parameters ...interface{}) {
	q := fmt.Sprintf(format, parameters...)
	if now {
		log.Println("EXEC: " + q)
		_, err := db.Exec(q)
		doErr(err)
	} else {
		q = reMidWhitespace.ReplaceAllString(q, " ")
		q = reTrailingWhitespace.ReplaceAllString(q, "")
		q = reLeadingWhitespace.ReplaceAllString(q, "")
		execString += fmt.Sprintf("%s;\n", q)
	}
}

func AddUnused(collection, column string) {
	unused = append(unused, collection+"."+column)
}

func SyncDb(db *sql.DB, model *databath.Model, now bool) {

	// CREATE DATABASE IF NOT EXISTS #{config.db.database}
	// USE #{config.db.database}
	// Probably won't work - the connection is set to a database.

	res, err := db.Query(`SHOW TABLE STATUS WHERE Engine != 'InnoDB'`)
	doErr(err)

	for res.Next() {
		table := TableStatus{}
		err := ScanToStruct(res, &table, "sql")
		doErr(err)
		MustExecF(now, db, "ALTER TABLE %s ENGINE = 'InnoDB'", table.Name)
	}
	res.Close()

	for collectionName, collection := range model.Collections {
		log.Printf("COLLECTION: %s\n", collectionName)
		//if collectionName[0:1] == "_" {
		//	log.Println("Skip super class table")
		//	continue
		//}
		res, err := db.Query(`SHOW TABLE STATUS WHERE Name = ?`, collectionName)
		doErr(err)
		if res.Next() {
			indexRes, err := db.Query(`
SELECT
  c.CONSTRAINT_NAME,
  c.TABLE_NAME,
  c.CONSTRAINT_TYPE,
  k.COLUMN_NAME,
  k.REFERENCED_TABLE_NAME,
  k.REFERENCED_COLUMN_NAME 
FROM information_schema.TABLE_CONSTRAINTS c
LEFT JOIN information_schema.KEY_COLUMN_USAGE k 
  ON c.CONSTRAINT_NAME = k.CONSTRAINT_NAME 
  AND c.TABLE_SCHEMA = k.TABLE_SCHEMA 
  AND c.TABLE_NAME = k.TABLE_NAME 
WHERE c.TABLE_SCHEMA = DATABASE() AND c.TABLE_NAME = "` + collectionName + `";
`)

			doErr(err)
			indexes := []*Index{}
			for indexRes.Next() {
				index := Index{}
				err := ScanToStruct(indexRes, &index, "sql")
				doErr(err)
				indexes = append(indexes, &index)
			}

			deferredStatements := []string{}

			allColumnsRes, err := db.Query(`SHOW COLUMNS FROM ` + collectionName)
			doErr(err)
			allColumns := map[string]Column{}
			for allColumnsRes.Next() {
				col := Column{}
				err := ScanToStruct(allColumnsRes, &col, "sql")
				doErr(err)
				allColumns[col.Field] = col
				has := false
				for colName, _ := range collection.Fields {
					if col.Field == colName {
						has = true
						break
					}

				}
				if !has {
					log.Printf("UNUSED COLUMN: %s.%s\n", collectionName, col.Field)
					AddUnused(collectionName, col.Field)
				}
			}
			allColumnsRes.Close()

			for colName, field := range collection.Fields {
				col, ok := allColumns[colName]
				if ok {
					colStr := col.GetString()
					modelStr := field.GetMysqlDef()
					if colStr != modelStr {
						log.Printf("CHANGE: '%s' '%s'\n", colStr, modelStr)

						// If VARCHAR(100) etc
						if reCheckLength.MatchString(modelStr) {
							matches := reCheckLength.FindStringSubmatch(modelStr)
							lenNewMax, _ := strconv.ParseUint(matches[1], 10, 64)
							var lenExists *uint64

							lengthRes, err := db.Query(fmt.Sprintf(`SELECT MAX(LENGTH(%s)) FROM %s`, colName, collectionName))
							if err != nil {
								doErr(err)
							} else {
								lengthRes.Next()
								err = lengthRes.Scan(&lenExists)
								lengthRes.Close()
								if err != nil {
									doErr(err)
								}
								if lenExists != nil {
									log.Printf("%s.%s max length: %d fits into %s\n",
										collectionName, colName, *lenExists, modelStr)
									if *lenExists > lenNewMax {
										doErr(fmt.Errorf("%s.%s len = %d, larger than new def %s",
											collectionName, colName, *lenExists, modelStr))
									}
								}
							}

						}
						MustExecF(now, db, "ALTER TABLE %s CHANGE COLUMN %s %s %s",
							collectionName, colName, colName, modelStr)
					}
				} else {
					MustExecF(now, db, "ALTER TABLE %s ADD `%s` %s", collectionName, colName, field.GetMysqlDef())
				}

				var linkToCollectionPtr *string

				refField, ok := field.Impl.(*types.FieldRef)
				if ok {
					linkToCollectionPtr = &refField.Collection
				} else {
					refIdField, ok := field.Impl.(*types.FieldRefID)
					if ok {
						linkToCollectionPtr = &refIdField.Collection
					}
				}
				if linkToCollectionPtr != nil {
					linkToCollection := *linkToCollectionPtr
					var matchedIndex *Index = nil
					for _, index := range indexes {
						// If Matches

						if index.ReferencedTableName != nil && colName == *index.ColumnName && *index.ReferencedTableName == linkToCollection {
							matchedIndex = index
							matchedIndex.Used = true
							break
						}

					}

					if matchedIndex == nil {
						// Create It.
						// Is it creatable with the current data?
						badRowsRes, err := db.Query(fmt.Sprintf(`SELECT id, %s FROM %s WHERE %s IS NOT NULL AND %s NOT IN (SELECT %s FROM %s)`, colName, collectionName, colName, colName, "id", linkToCollection))
						if err != nil {
							log.Printf("Error on FK check for %s.%s\n", collectionName, colName)

						} else {

							hasBad := false
							for badRowsRes.Next() {
								hasBad = true
								id := 0
								fkVal := 0
								badRowsRes.Scan(&id, &fkVal)
								log.Printf("Foreign Key Test Fail: Entry %d for %s.%s references %s.id = %d, which doesn't exist\n", id, collectionName, colName, linkToCollection, fkVal)
							}
							badRowsRes.Close()
							if hasBad {
								panic("Foreign Key Failure, see above.")
							}

							deferredStatements = append(deferredStatements, fmt.Sprintf(`ALTER TABLE %s 
								ADD CONSTRAINT fk_%s_%s_%s_%s 
								FOREIGN KEY (%s) 
								REFERENCES %s(%s)`, collectionName, collectionName, colName, linkToCollection, "id", colName, linkToCollection, "id"))
						}

					}
				}
			}
			for _, index := range indexes {
				if !index.Used && *index.ConstraintType == "FOREIGN KEY" {
					MustExecF(now, db, `
					ALTER TABLE %s DROP FOREIGN KEY %s`, collectionName, index.ConstraintName)
				}
			}
			for _, statement := range deferredStatements {
				MustExecF(now, db, statement)
			}

		} else {
			// CREATE!
			params := make([]string, 0, 0)

			for name, field := range collection.Fields {
				params = append(params, fmt.Sprintf("`%s` %s", name, field.GetMysqlDef()))
			}

			params = append(params, "PRIMARY KEY (`id`)")

			MustExecF(now, db, "CREATE TABLE %s (%s)", collectionName, strings.Join(params, ", "))
		}
		res.Close()
	}

	log.Println("==========")
	log.Printf("\n\n%s\n\n", execString)
	log.Printf("Unused: \n%s\n", strings.Join(unused, "\n"))
}
