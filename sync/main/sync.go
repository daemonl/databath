package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/daemonl/databath"
	"github.com/daemonl/databath/sync"
	_ "github.com/go-sql-driver/mysql"
)

var modelFilename string
var dsn string
var force bool

func init() {
	flag.StringVar(&modelFilename, "model", "", "the data model json file to read")
	flag.StringVar(&dsn, "dsn", "", "the golang style sql connection string")
	flag.BoolVar(&force, "force", false, "run the update commands NOW")
}

func main() {
	flag.Parse()
	if len(modelFilename) < 1 {
		fmt.Println("No model filename specified (use --model)")
		os.Exit(1)
		return
	}

	model, err := databath.ReadModelFromFileForSync(modelFilename)
	if err != nil {
		fmt.Printf("Error reading model: %s\n", err.Error())
		os.Exit(2)
		return
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("Error opening database %s: %s\n", dsn, err.Error())
		os.Exit(3)
		return
	}
	defer db.Close()
	sync.SyncDb(db, model, force)
}
