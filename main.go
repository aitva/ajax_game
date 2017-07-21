package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

type server struct {
	tmpl *template.Template
}

func main() {
	var err error
	s := &server{}
	s.tmpl, err = template.ParseGlob("template/*.html")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", s.indexHandler)
	logHandler := handlers.LoggingHandler(os.Stdout, mux)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", logHandler))
}

func (s *server) indexHandler(w http.ResponseWriter, r *http.Request) {
	err := s.tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Println("fail to execute template:", err)
	}
}
