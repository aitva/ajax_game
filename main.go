package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"net/http/httputil"

	"github.com/gorilla/handlers"
)

type server struct {
	tmpl *template.Template
}

type displayHTTP struct {
	Title   string
	Icon    string
	Request string
}

func main() {
	var err error

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	s := &server{}
	s.tmpl, err = parseTemplates("template", ".html")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.indexHandler)
	mux.HandleFunc("/register/", s.registerHandler)
	mux.HandleFunc("/template/reload/", s.reloadTemplate)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// f, err := os.Create("access.log")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	logHandler := handlers.LoggingHandler(os.Stdout, mux)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", logHandler))
}

func writeError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
}

func (s *server) indexHandler(w http.ResponseWriter, r *http.Request) {
	d := &displayHTTP{
		Title: "AJAX Game",
		Icon:  "fa-gamepad",
	}
	err := s.tmpl.ExecuteTemplate(w, "index.html", d)
	if err != nil {
		log.Println("fail to execute template:", err)
		writeError(w, "Oops, something went wrong... ☹")
		return
	}
}

func (s *server) reloadTemplate(w http.ResponseWriter, r *http.Request) {
	var err error
	s.tmpl, err = parseTemplates("template", ".html")
	if err != nil {
		log.Println("fail to parse template:", err)
		writeError(w, "Oops, something went wrong... ☹")
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("Ready to go!"))
}

func (s *server) registerHandler(w http.ResponseWriter, r *http.Request) {
	d := &displayHTTP{
		Title: "Register",
		Icon:  "fa-address-card",
	}
	tmp, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println("fail to dump request:", err)
		writeError(w, "Oops, something went wrong... ☹")
		return
	}
	d.Request = string(tmp)
	err = s.tmpl.ExecuteTemplate(w, "register.html", d)
	if err != nil {
		log.Println("fail to execute template:", err)
		writeError(w, "Oops, something went wrong... ☹")
		return
	}
}
