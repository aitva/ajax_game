package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
)

type jsonResponse struct {
	Title   string        `json:"title"`
	Icon    string        `json:"icon"`
	Text    template.HTML `json:"text"`
	Editor  bool          `json:"editor"`
	Objects []GameObject  `json:"objects"`
}

type gameServer struct {
	tmpls *template.Template
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	tmpls, err := template.ParseGlob("template/*.html")
	if err != nil {
		log.Fatal("fail to parse template:", err)
	}
	server := gameServer{
		tmpls: tmpls,
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.Handle("/", server.pageHandler("index"))
	mux.Handle("/freedom/", server.pageHandler("freedom"))
	mux.Handle("/name/", server.pageHandler("name"))
	mux.Handle("/settings/", server.pageHandler("settings"))
	mux.Handle("/super8/", server.pageHandler("super8"))

	logHandler := handlers.LoggingHandler(os.Stdout, mux)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", logHandler))
}

func (s *gameServer) pageHandler(page string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		page, err := parsePage(page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		meta, err := page.Meta()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		locked := false

		// If the page requires objects
		if len(meta.Required) > 0 {
			// Check if the user has objects to unlock it
			usedObjects := getObjects(r)
			locked = isLocked(meta.Required, usedObjects)
		}

		name := getName(r)

		content, err := page.Content(name, locked)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := jsonResponse{
			Icon:    meta.Icon,
			Title:   meta.Title,
			Editor:  meta.Editor,
			Objects: meta.Discovered,
			Text:    template.HTML(content),
		}
		if r.Header.Get("Accept") == "application/json" {
			respond(w, resp)
			return
		}
		err = s.tmpls.ExecuteTemplate(w, "index.html", resp)
		if err != nil {
			log.Println("fail to execute template:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func parsePage(pageName string) (Page, error) {
	f, err := os.Open("pages/" + pageName + ".md")
	if err != nil {
		return nil, err
	}

	page := &page{}
	err = page.Parse(f)
	if err != nil {
		return nil, err
	}

	return page, nil
}

func respond(w http.ResponseWriter, resp jsonResponse) {
	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// getObjects extracts object from a request.
// It expects object to be in the form: key=value; key2=value2; ...
func getObjects(r *http.Request) []GameObject {
	objects := make([]GameObject, 0)

	header := r.Header.Get("Use-Object")
	if header == "" {
		return objects
	}

	tokens := strings.Split(header, ";")
	for _, t := range tokens {
		exploded := strings.Split(t, "=")
		if len(exploded) != 2 {
			continue
		}
		name := strings.TrimSpace(exploded[0])
		value := strings.TrimSpace(exploded[1])

		o := GameObject{name, value}
		objects = append(objects, o)
	}

	return objects
}

func isLocked(locks, usedObjects []GameObject) bool {

	if len(locks) == 0 {
		return false
	}

	if len(usedObjects) == 0 {
		return true
	}

lock:
	for _, lock := range locks {
		for _, used := range usedObjects {
			if used.Name == lock.Name && used.Value == lock.Value {
				continue lock
			} else {
				return true
			}
		}
	}

	return false
}

func getName(r *http.Request) string {
	return r.Header.Get("Name")
}
