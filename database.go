package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

// database holds the global database connection handler.
var database *sql.DB

// intializeDatabase sets up our application's database connection.
func initializeDatabase() {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		getenv("DB_USER", ""),
		getenv("DB_PASS", ""),
		getenv("DB_HOST", ""),
		getenv("DB_PORT", ""),
		getenv("DB_NAME", ""),
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

	maybeCreateTable()
}

// maybeCreateTable creates our database table if it doesn't exist.
//
// If this application grows any larger it may be desirable to replace
// this behavior with a proper ORM (there are a few decent options in
// the Go ecosystem.)
func maybeCreateTable() {
	stmt, err := database.Prepare(`CREATE TABLE IF NOT EXISTS ` + "`requests`" + ` (
		` + "`ID`" + ` int(11) NOT NULL AUTO_INCREMENT,
		` + "`domain`" + ` varchar(255) NOT NULL,
		` + "`path`" + ` varchar(255) NOT NULL,
		` + "`user`" + ` varchar(255) NOT NULL,
		` + "`timezone`" + ` varchar(255) NOT NULL,
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
