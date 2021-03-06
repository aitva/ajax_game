package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"text/template"

	"github.com/russross/blackfriday"
	"gopkg.in/yaml.v2"
)

// GameObject represents an object found or used by the player
type GameObject struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// PageMeta holds page meta informations parsed from the Markdown pages
type PageMeta struct {
	Icon       string       `yaml:"icon"`
	Title      string       `yaml:"title"`
	Editor     bool         `yaml:"editor"`
	Required   []GameObject `yaml:"required"`
	Discovered []GameObject `yaml:"discovered"`
}

// Page represents all informations form a Markdown page
type Page interface {
	Meta() (*PageMeta, error)
	Content(name string, locked bool) (string, error)

	Parse(r io.Reader) error
}

// page implements the Page interface
type page struct {
	frontMatter []byte
	text        []byte
}

const (
	frontMatterDelimiter = "```"
)

func (p *page) Parse(r io.Reader) error {
	reader := bufio.NewReader(r)

	peek, err := reader.Peek(3)
	if err != nil {
		return err
	}

	if string(peek) != frontMatterDelimiter {
		return fmt.Errorf("no opening FrontMatter delimiter found: %s instead", peek)
	}

	reader.ReadLine()

	frontMatter, err := readUntil(reader, []byte(frontMatterDelimiter))
	if err != nil {
		return err
	} else if err == io.EOF {
		return errors.New("did not find closing FrontMatter delimiter")
	}

	text, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	p.frontMatter = frontMatter
	p.text = text

	return nil
}

type reader interface {
	ReadString(delim byte) (line string, err error)
}

func readUntil(r reader, delim []byte) (line []byte, err error) {
	for {
		s := ""
		s, err = r.ReadString(delim[len(delim)-1])
		if err != nil {
			return
		}

		line = append(line, []byte(s)...)

		if bytes.HasSuffix(line, delim) {
			return line[:len(line)-len(delim)], nil
		}
	}
}

func (p *page) Meta() (*PageMeta, error) {
	meta := &PageMeta{}

	err := yaml.Unmarshal(p.frontMatter, meta)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

func (p *page) Content(name string, locked bool) (string, error) {
	data := struct {
		Name   string
		Locked bool
	}{
		Name:   name,
		Locked: locked,
	}

	t := template.Must(template.New("content").Parse(string(p.text)))

	var out bytes.Buffer
	err := t.Execute(&out, data)
	if err != nil {
		return "", err
	}

	str := blackfriday.Run(out.Bytes())
	return string(str), nil
}
