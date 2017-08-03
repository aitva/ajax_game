package main

import (
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"reflect"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/handlers"
)

type server struct {
	tmpl *template.Template
}

type headInfo struct {
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
	mux.HandleFunc("/http-media/", s.httpMediaHandler)
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

func (s *server) execTemplate(w http.ResponseWriter, file string, data interface{}) {
	err := s.tmpl.ExecuteTemplate(w, file, data)
	if err != nil {
		log.Println("fail to execute template:", err)
		writeError(w, "Oops, something went wrong... ☹")
		return
	}
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
	d := &headInfo{
		Title: "AJAX Game",
		Icon:  "fa-gamepad",
	}
	s.execTemplate(w, "index.html", d)
}

func (s *server) httpBasicsHandler(file, title, icon string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d := &headInfo{
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
		s.execTemplate(w, file, d)
	}
}

func (s *server) httpPostHandler(w http.ResponseWriter, r *http.Request) {
	d := &struct {
		headInfo
		Error string
		Color string
		Name  string
	}{
		headInfo: headInfo{
			Title: "HTTP Post",
			Icon:  "fa-info",
		},
		Error: "Sorry, I am confuse... What is your favorite color again? Did we met before?",
	}
	badRequestHandler := s.badRequestHandler(d)

	q := r.URL.Query()
	d.Name = strings.Title(q.Get("name"))
	d.Color = q.Get("color")
	if d.Name == "" || d.Color == "" {
		badRequestHandler(w, r)
		return
	}

	tmp, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println("fail to dump request:", err)
		writeError(w, "Oops, something went wrong... ☹")
		return
	}
	d.Request = string(tmp)
	s.execTemplate(w, "http-post.html", d)
}

func (s *server) httpMediaHandler(w http.ResponseWriter, r *http.Request) {
	d := &struct {
		headInfo
		Error    string
		Color    string
		Name     string
		KindWord string
		FunGauge string
	}{
		headInfo: headInfo{
			Title: "HTTP Media",
			Icon:  "fa-sun-o",
		},
		Error: "Something is not quite right... Do we know each other?",
	}
	badRequestHandler := s.badRequestHandler(d)

	tmp, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println("fail to dump request:", err)
		writeError(w, "Oops, something went wrong... ☹")
		return
	}
	d.Request = string(tmp)

	err = r.ParseForm()
	if err != nil {
		badRequestHandler(w, r)
		return
	}
	d.Color = r.Form.Get("color")
	d.Name = strings.Title(r.Form.Get("name"))
	d.KindWord = r.Form.Get("kind-word")
	d.FunGauge = r.Form.Get("fun-gauge")
	if d.Color == "" || d.Name == "" || d.KindWord == "" || d.FunGauge == "" {
		badRequestHandler(w, r)
		return
	}

	s.execTemplate(w, "http-media.html", d)
}

func (s *server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	d := &headInfo{
		Title: "404 - Not Found",
		Icon:  "fa-eye",
	}
	w.WriteHeader(http.StatusNotFound)
	err := s.tmpl.ExecuteTemplate(w, "not-found.html", d)
	if err != nil {
		log.Println("fail to execute template:", err)
		writeError(w, "Oops, something went wrong... ☹")
		return
	}
}

func (s *server) badRequestHandler(data interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := reflect.ValueOf(data)
		if v.Kind() == reflect.Ptr && !v.IsNil() && v.IsValid() {
			f := v.Elem().FieldByName("Icon")
			if f.IsValid() {
				f.SetString("fa-bug")
			}
		} else {
			log.Println("unexpected type for data (expects &struct{})")
		}

		w.WriteHeader(http.StatusBadRequest)
		err := s.tmpl.ExecuteTemplate(w, "bad-request.html", data)
		if err != nil {
			log.Println("fail to execute template:", err)
			writeError(w, "Oops, something went wrong... ☹")
			return
		}
	}
}
