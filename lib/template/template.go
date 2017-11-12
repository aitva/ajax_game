package template

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// URL represents a raw URL.
type URL = template.URL

// HTML contains raw HTML.
type HTML = template.HTML

// Template represents an HTML template with partial support
// and hot reload.
type Template struct {
	folder string
	ext    string
	m      sync.RWMutex
	stop   chan struct{}
	tmpls  *template.Template
}

// New instanciates a new Template.
func New(folder, ext string) (*Template, error) {
	t := &Template{
		folder: folder,
		ext:    ext,
		stop:   make(chan struct{}),
	}
	err := t.load("")
	return t, err
}

// ExecuteTemplate executes a template an write the output on w.
func (t *Template) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	t.m.RLock()
	err := t.tmpls.ExecuteTemplate(w, name, data)
	t.m.RUnlock()
	return err
}

// Lookup searches for a named template.
// It returns true if the template if found, false otherwise.
func (t *Template) Lookup(name string) bool {
	t.m.RLock()
	tmpl := t.tmpls.Lookup(name)
	t.m.RUnlock()
	return tmpl != nil
}

// load loads templates from a directory and its subdirectory.
// TODO: improve with filepath.Walk
func (t *Template) load(dir string) error {
	t.m.Lock()
	defer t.m.Unlock()
	t.tmpls = template.New("")
	return filepath.Walk(t.folder, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), t.ext) {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		data, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		name, _ := filepath.Rel(t.folder, path)
		name = name[:len(name)-len(t.ext)]
		_, err = t.tmpls.New(name).Parse(string(data))
		return err
	})
}

// Watch watches for change in the template folder.
func (t *Template) Watch() error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	err = filepath.Walk(t.folder, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return err
		}
		w.Add(path)
		return nil
	})
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case e := <-w.Events:
				if e.Name[0] == '.' || !strings.HasSuffix(e.Name, t.ext) {
					continue
				}
				if e.Op&fsnotify.Write != fsnotify.Write {
					continue
				}
				log.Println("watch event:", e)
				err := t.load("")
				if err != nil {
					log.Println("fail to load template:", err)
				}
			case err := <-w.Errors:
				log.Println("watch error:", err)
			case <-t.stop:
				t.stop <- struct{}{}
				return
			}
		}
	}()
	return nil
}

// StopWatch stops to watch for change in the template folder,
func (t *Template) StopWatch() error {
	// send stop signal
	t.stop <- struct{}{}
	// wait for response
	<-t.stop
	return nil
}
