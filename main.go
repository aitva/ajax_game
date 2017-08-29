package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

type gameObject struct {
	Name  string
	Value string
}

type jsonResponse struct {
	Title   string
	Icon    string
	Text    string
	Objects []*gameObject
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.Handle("/home", pageHandler("home"))
	mux.Handle("/locked/", pageHandler("locked"))
	mux.Handle("/closet/", pageHandler("closet"))

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	logHandler := handlers.LoggingHandler(os.Stdout, mux)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", logHandler))
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	resp := &jsonResponse{
		Title: "404 - Not Found",
		Icon:  "fa-eye",
	}
	w.WriteHeader(http.StatusNotFound)
	respond(w, resp)
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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		notFoundHandler(w, r)
		return
	}

	resp := &jsonResponse{
		Title: "AJAX Game",
		Icon:  "fa-gamepad",
		Text:  "Go to /home",
	}

	respond(w, resp)
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

func parsePage(page string) (*jsonResponse, error) {
	content, err := ioutil.ReadFile("pages/" + page + ".md")
	if err != nil {
		return nil, err
	}

	s := string(content)
	resp := &jsonResponse{
		Text: s,
	}

	return resp, nil
}
