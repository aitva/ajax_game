package main

import (
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/fsnotify/fsnotify"
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
	mux.HandleFunc("/http-basics/", s.httpBasicsHandler("http-basics.html", "HTTP Basics", "fa-coffee"))
	mux.HandleFunc("/http-query/", s.httpBasicsHandler("http-query.html", "HTTP Query", "fa-question"))
	mux.HandleFunc("/http-post/", s.httpPostHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// f, err := os.Create("access.log")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	logHandler := handlers.LoggingHandler(os.Stdout, mux)

	w, err := watchTemplate("template")
	defer w.Close()
	go s.reloadTemplate(w)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", logHandler))
}

func writeError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
}

func (s *server) reloadTemplate(w *fsnotify.Watcher) {
	for {
		select {
		case event := <-w.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				tmpl, err := parseTemplates("template", ".html")
				if err != nil {
					log.Println("fail to load template:", err)
				}
				log.Println("template updated")
				s.tmpl = tmpl
			}
		case err := <-w.Errors:
			log.Println("fail to watch template:", err)
		}
	}
}

func (s *server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.notFoundHandler(w, r)
		return
	}
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

func (s *server) httpBasicsHandler(file, title, icon string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d := &displayHTTP{
			Title: title,
			Icon:  icon,
		}
		tmp, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Println("fail to dump request:", err)
			writeError(w, "Oops, something went wrong... ☹")
			return
		}
		d.Request = string(tmp)
		err = s.tmpl.ExecuteTemplate(w, file, d)
		if err != nil {
			log.Println("fail to execute template:", err)
			writeError(w, "Oops, something went wrong... ☹")
			return
		}
	}
}

func (s *server) httpPostHandler(w http.ResponseWriter, r *http.Request) {
	d := struct {
		*displayHTTP
		Color string
		Name  string
	}{
		displayHTTP: &displayHTTP{
			Title: "HTTP Post",
			Icon:  "fa-info",
		},
	}
	q := r.URL.Query()
	d.Name = q.Get("name")
	d.Color = q.Get("color")
	if d.Name == "" || d.Color == "" {
		d.Icon = "fa-bug"
	}
	tmp, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println("fail to dump request:", err)
		writeError(w, "Oops, something went wrong... ☹")
		return
	}
	d.Request = string(tmp)

	err = s.tmpl.ExecuteTemplate(w, "http-post.html", d)
	if err != nil {
		log.Println("fail to execute template:", err)
		writeError(w, "Oops, something went wrong... ☹")
		return
	}
}

func (s *server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	d := &displayHTTP{
		Title: "404 - Not Found",
		Icon:  "fa-eye",
	}
	err := s.tmpl.ExecuteTemplate(w, "not-found.html", d)
	if err != nil {
		log.Println("fail to execute template:", err)
		writeError(w, "Oops, something went wrong... ☹")
		return
	}
}
