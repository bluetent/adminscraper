package main

import (
	"crypto/tls"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/urfave/negroni"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"os"
)

func main() {
	readConfig()

	initializeDatabase()
	defer database.Close()

	handler := initializeHandler()
	port := getenv("PORT", "80")
	httpsPort := getenv("HTTPSPORT", "443")
	domain := getenv("DOMAIN", "gollector.bluetent.com")

	log.Printf("Serving HTTP at %s:%s", domain, port)
	go http.ListenAndServe(":"+port, http.HandlerFunc(redirectToHTTPS))

	log.Printf("Serving HTTPS at %s:%s", domain, httpsPort)

	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache("certs"),
	}

	server := &http.Server{
		Addr: ":" + httpsPort,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
		Handler: handler,
	}

	server.ListenAndServeTLS("", "")
}

// redirectToHTTPS sends the request on to the https:// version of the URL.
// Snagged from https://gist.github.com/d-schmidt/587ceec34ce1334a5e60
func redirectToHTTPS(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	target := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target,
		// see @andreiavrammsd comment: often 307 > 301
		http.StatusTemporaryRedirect)
}

func initializeHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", logHit)

	n := negroni.Classic()
	n.Use(negroni.HandlerFunc(addCORSHeaders))
	n.Use(negroni.HandlerFunc(respondOptions))
	n.Use(negroni.HandlerFunc(rejectNonPOSTRequests))
	n.UseHandler(mux)
	return n
}

func readConfig() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading environment")
	}
}

func addCORSHeaders(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Access-Control-Allow-Origin")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	next(w, r)
}

func respondOptions(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Method == "OPTIONS" {
		return
	}
	next(w, r)
}

func rejectNonPOSTRequests(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid request.")
		return
	}
	next(w, r)
}

func logHit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	domain := r.FormValue("domain")
	path := r.FormValue("path")
	user := r.FormValue("user")
	timezone := r.FormValue("timezone")
	address := r.RemoteAddr

	stmt, err := database.Prepare(`INSERT INTO requests SET
	      domain=?,
				path=?,
				user=?,
				timezone=?,
				address=?,
				created=NOW()
	`)
	if err != nil {
		log.Panic("Error preparing insert statement")
		log.Panic(err)
	}

	err = executeStatement(stmt, domain, path, user, timezone, address)

	if err != nil {
		log.Panic(err)
		return
	}

	fmt.Fprintf(w, "Logged: %v+", domain, path, user, timezone, address)
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
