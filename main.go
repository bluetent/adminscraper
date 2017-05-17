package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var database *sql.DB

func main() {
	config := readConfig()
	log.SetOutput(os.Stderr)

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		config.User,
		config.Pass,
		config.Host,
		config.Port,
		config.Database,
	))

	if err != nil {
		log.Panic("Error in MySQL parameters")
		log.Panic(err)
	}

	err = db.Ping()
	if err != nil {
		log.Panic("Error opening MySQL connection")
		log.Panic(err)
	}
	database = db
	defer database.Close()

	maybeSetupDatabase()

	http.HandleFunc("/", logHit)
	port := getenv("PORT", "8000")
	log.Printf("Serving on port %s", port)
	http.ListenAndServe(":"+port, nil)
}

type config struct {
	Host     string
	Port     string
	User     string
	Pass     string
	Database string
}

func readConfig() (cfg config) {
	data, err := ioutil.ReadFile("config.yml")

	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(data, &cfg)

	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

func maybeSetupDatabase() {
	stmt, err := database.Prepare(`CREATE TABLE IF NOT EXISTS ` + "`requests`" + ` (
		` + "`ID`" + ` int(11) NOT NULL AUTO_INCREMENT,
		` + "`domain`" + ` varchar(255) NOT NULL,
		` + "`path`" + ` varchar(255) NOT NULL,
		` + "`user`" + ` varchar(255) NOT NULL,
		` + "`address`" + ` varchar(255) NOT NULL,
		` + "`created`" + ` time NOT NULL,
		PRIMARY KEY (` + "`ID`" + `)
	) ENGINE=InnoDB  DEFAULT CHARSET=latin1 AUTO_INCREMENT=3;`)

	if err != nil {
		log.Panic("Error preparing table create statement")
		log.Panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		log.Panic("Error creating table")
		log.Panic(err)
	}
}

func logHit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprint(w, "Invalid request.")
		return
	}

	r.ParseForm()
	domain := r.FormValue("domain")
	path := r.FormValue("path")
	user := r.FormValue("user")
	address := r.RemoteAddr

	stmt, err := database.Prepare("INSERT INTO requests SET domain=?, path=?, user=?, address=?, created=NOW()")
	if err != nil {
		log.Panic("Error preparing insert statement")
		log.Panic(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(domain, path, user, address)
	if err != nil {
		log.Panic("Error executing insert statement")
		log.Panic(err)
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		log.Panic("Error getting ID")
		log.Panic(err)
	}

	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Panic("Error getting row count.")
		log.Panic(err)
	}
	log.Printf("ID = %d, affected = %d\n", lastId, rowCnt)

	fmt.Fprintf(w, "Logged: %s, %s, %s, %s", domain, path, user, address)
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
