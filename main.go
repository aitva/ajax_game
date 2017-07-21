package main

import (
	"log"
	"net/http"

	"os"

	"strings"

	"github.com/gorilla/handlers"
)

func main() {
	log.Println("listening on :8080")
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	logHandler := handlers.LoggingHandler(os.Stdout, mux)
	log.Fatal(http.ListenAndServe(":8080", logHandler))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	headers := ""
	for k, v := range r.Header {
		headers += k + ": " + strings.Join(v, ",") + "\n"
	}
	w.Write([]byte("HEADERS:\n"))
	w.Write([]byte(headers + "\n"))

	requestedWith := r.Header.Get("HTTP_X_REQUESTED_WITH")
	if strings.ToLower(requestedWith) == "xmlhttprequest" {
		w.Write([]byte("Hello AJAX!"))
		return
	}
	w.Write([]byte("Hello World!"))
}
