package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"os"
)

func main() {
	readConfig()
	log.SetOutput(os.Stderr)

	initializeDatabase()

	handler := initializeHandler()
	port := getenv("PORT", "8000")
	log.Printf("Serving on port %s", port)
	http.ListenAndServe(":"+port, handler)
}

func initializeHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", logHit)

	n := negroni.Classic()
	n.UseHandler(mux)
	return n
}

func readConfig() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading environment")
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
