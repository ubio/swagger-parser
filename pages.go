package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v1"
)

// Return a list of markdown pages for the API
func getPages() []Page {
	pages := Pages{}

	content, err := ioutil.ReadFile(pagesFile)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(content, &pages)
	if err != nil {
		log.Fatal(err)
	}

	return pages.Pages
}

type Pages struct {
	Pages []Page `yaml:"pages"`
}

type Page struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Filename    string `yaml:"filename"`
	Paths       []Path `yaml:"paths"`
}

type Path struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
}
