package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aitva/ajax_game/lib/template"
	"github.com/gorilla/handlers"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	tmpls, err := template.New("template", ".html")
	if err != nil {
		log.Fatal("fail to parse template:", err)
	}
	err = tmpls.Watch()
	if err != nil {
		log.Fatal("fail to watch template:", err)
	}
	server := GameServer{
		tmpls: tmpls,
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("/", server.mainHandler)
	mux.Handle("/freedom/", server.pageHandler("freedom"))
	mux.Handle("/name/", server.pageHandler("name"))
	mux.Handle("/settings/", server.pageHandler("settings"))
	mux.Handle("/super8/", server.pageHandler("super8"))

	logHandler := handlers.LoggingHandler(os.Stdout, mux)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", logHandler))
}

// PageView contains mandatory data to render a page.
type PageView struct {
	Title   string        `json:"title"`
	Icon    string        `json:"icon"`
	Text    template.HTML `json:"text"`
	Editor  bool          `json:"editor"`
	Objects []GameObject  `json:"objects"`
}

// GameServer represents the main game server.
type GameServer struct {
	tmpls *template.Template
}

func (s *GameServer) executeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println("fail to execute JSON:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *GameServer) executeTemplate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.tmpls.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Println("fail to execute template:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *GameServer) mainHandler(w http.ResponseWriter, r *http.Request) {
	name := "index"
	if r.URL.Path != "/" {
		name = "404"
	}
	s.executeTemplate(w, name, nil)
}

func (s *GameServer) pageHandler(page string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+page+"/" {
			s.executeTemplate(w, "404", nil)
			return
		}

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

		resp := PageView{
			Icon:    meta.Icon,
			Title:   meta.Title,
			Editor:  meta.Editor,
			Objects: meta.Discovered,
			Text:    template.HTML(content),
		}
		if r.Header.Get("Accept") == "application/json" {
			s.executeJSON(w, resp)
			return
		}
		s.executeTemplate(w, "page", resp)
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
