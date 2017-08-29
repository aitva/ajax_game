package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
)

type jsonResponse struct {
	Title   string        `json:"title"`
	Icon    string        `json:"icon"`
	Text    string        `json:"text"`
	Objects []*GameObject `json:"objects"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("static")))

	mux.Handle("/home", pageHandler("home"))
	mux.Handle("/locked/", pageHandler("locked"))
	mux.Handle("/closet/", pageHandler("closet"))

	logHandler := handlers.LoggingHandler(os.Stdout, mux)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", logHandler))
}

func pageHandler(page string) http.Handler {
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

		content, err := page.Content(locked)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := &jsonResponse{
			Icon:    meta.Icon,
			Title:   meta.Title,
			Objects: meta.Discovered,
			Text:    content,
		}

		respond(w, resp)
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

func respond(w http.ResponseWriter, resp *jsonResponse) {
	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func getObjects(r *http.Request) []*GameObject {
	objects := make([]*GameObject, 0)

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

		o := &GameObject{name, value}
		objects = append(objects, o)
	}

	return objects
}

func isLocked(locks, usedObjects []*GameObject) bool {

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
