package main

import (
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func parseTemplates(folder, ext string) (*template.Template, error) {
	if folder[len(folder)-1] != filepath.Separator {
		folder += string(filepath.Separator)
	}
	t := template.New("")
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if !strings.Contains(path, ".html") {
			return nil
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		s := string(b)
		filename := path[len(folder):]
		tmpl := t.New(filename)
		_, err = tmpl.Parse(s)
		return err
	})
	return t, err
}
