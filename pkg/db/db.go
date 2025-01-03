package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

const connStrCfg = "%s:%s@tcp(%s:%s)/server_cfg"
const connStrCodes = "%s:%s@tcp(%s:%s)/codes"

var DBCfg *sql.DB
var DBCodes *sql.DB

var initialized bool = false

func init() {
	if initialized {
		return
	}

	// connect to DB and set objects
	user := os.Getenv("db_user")
	pass := os.Getenv("db_pass")
	host := os.Getenv("db_host")
	port := os.Getenv("db_port")

	var err error // bc DBs are already initialized
	log.Println("Initializing server config db...")
	DBCfg, err = sql.Open("mysql", fmt.Sprintf(connStrCfg, user, pass, host, port))
    if err != nil {
        log.Fatal(err)
    }
	log.Println("Server config db initialized!")

	log.Println("Initializing codes db...")
	DBCodes, err = sql.Open("mysql", fmt.Sprintf(connStrCodes, user, pass, host, port))
    if err != nil {
        log.Fatal(err)
    }
	log.Println("Codes db initialized!")

	initialized = true
}