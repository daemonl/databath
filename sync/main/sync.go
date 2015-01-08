package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/daemonl/databath"
	"github.com/daemonl/databath/sync"
	"github.com/daemonl/go_gsd/torch"
	_ "github.com/go-sql-driver/mysql"
)

var modelFilename string
var dsn string
var force bool
var setUser string

func init() {
	flag.StringVar(&modelFilename, "model", "", "the data model json file to read")
	flag.StringVar(&dsn, "dsn", "", "the golang style sql connection string")
	flag.BoolVar(&force, "force", false, "run the update commands NOW")
	flag.StringVar(&setUser, "setuser", "", "specify a user for development username:password, will create or update")
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
	mig, err := sync.BuildMigration(db, model)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
		return
	}

	e, err := json.Marshal(mig)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
		return
	}
	fmt.Println(string(e))

	if force {
		err := mig.Run(db)
		if err != nil {
			fmt.Println(err)
			os.Exit(5)
		}
	}

	if len(setUser) > 0 {
		parts := strings.Split(setUser, ":")
		if len(parts) != 2 {
			fmt.Fprintln(os.Stderr, "setuser must be in the form username:password")
			os.Exit(1)
		}

		err = SetUser(db, "user", parts[0], parts[1])
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(5)
			return
		}
	}
}

func SetUser(db *sql.DB, table string, username string, password string) error {
	log.Println("SET USER")

	rows, err := db.Query(`SELECT id FROM `+table+` WHERE username = ?`, username)
	if err != nil {
		return err
	}
	var id uint64 = 0
	if rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			log.Println(err)
		}
		log.Printf("Has Next %d\n", id)
	}
	rows.Close()

	hashedPassword := torch.HashPassword(password)

	if id != 0 {

		_, err := db.Exec(`UPDATE `+table+` SET password = ? WHERE id = ?`, hashedPassword, id)
		if err != nil {
			return err
		}

	} else {
		_, err := db.Exec(`INSERT INTO `+table+` (username, password, set_on_next_login, access) VALUES (?, ?, 0, 0)`, username, hashedPassword)

		if err != nil {
			return err
		}

	}
	return nil
}
