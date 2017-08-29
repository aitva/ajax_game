package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

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
		resp, err := parsePage(page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		respond(w, resp)
	})
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

func parsePage(pageName string) (*jsonResponse, error) {
	f, err := os.Open("pages/" + pageName + ".md")
	if err != nil {
		return nil, err
	}

	page := &page{}
	page.Parse(f)

	resp := &jsonResponse{
		Title: page.Meta().Title,
		Text:  page.Content(),
	}

	return resp, nil
}
