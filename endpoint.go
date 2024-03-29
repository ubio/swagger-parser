package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ubio/swagger-parser/pkg/curl"
)

type Param struct {
	Name        string      `json:"name"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Format      string      `json:"format"`
	Example     string      `json:"example"`
	EnumJSON    string      `json:"enum_json"`
	MaxItems    *float64    `json:"max_items"`
	MinItems    *float64    `json:"min_items"`
	Default     interface{} `json:"default"`
}

type Example struct {
	Value string `json:"value"`
}

type ResponseExamples []ResponseExample

type ResponseExample struct {
	Key     string
	Value   string
	Summary string
	Status  int
}

type Server struct {
	URL string `json:"url"`
}

type RequestParams []RequestParam

type RequestParam struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Format      string `json:"format"`
	Example     string `json:"example"`
	ExampleJSON interface{}
	Enum        []string    `json:"enum"`
	EnumJSON    string      `json:"enum_json"`
	Required    bool        `json:"required"`
	MaxItems    *float64    `json:"max_items"`
	MinItems    *float64    `json:"min_items"`
	Default     interface{} `json:"default"`
}

type Endpoint struct {
	Server                string
	Path                  string
	Method                string
	Info                  *openapi3.Operation
	Params                map[string][]Param
	QueryParams           []Param
	HeaderParams          []Param
	PathParams            []Param
	Curl                  string
	RequestParams         RequestParams
	RequestExamples       RequestExamples
	ResponseExamples      ResponseExamples // this is rendered in a slot
	ResponseErrorExamples ResponseExamples
	ResponseExampleKeys   string // this is passed to Vue as a csv
	Examples              map[string]string
}

func (e *Endpoint) createExample(lang string) {
	example := ""
	if e.Examples == nil {
		e.Examples = make(map[string]string)
	}

	switch lang {
	case "curl":
		c := curl.NewCommand(e.Server, e.Method, e.pathExample(), e.exampleHeaders(), e.exampleQueryParams(), e.exampleRequestBody())
		example = c.ExampleString
	}

	e.Examples[lang] = example
}

// substitute the path params `/path/:id/resource` with the example value
func (e Endpoint) pathExample() string {
	parts := strings.Split(e.Path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			for _, param := range e.PathParams {
				if fmt.Sprintf(":%s", param.Name) == part {
					parts[i] = param.Example
				}
			}
		}
	}
	return strings.Join(parts, "/")
}

func (e Endpoint) exampleHeaders() []string {
	cp := make([]string, len(e.HeaderParams))
	for _, p := range e.HeaderParams {
		cp = append(cp, p.Example)
	}
	return cp
}

func (e Endpoint) exampleQueryParams() []string {
	cp := make([]string, len(e.QueryParams))
	for _, p := range e.QueryParams {
		cp = append(cp, p.Example)
	}
	return cp
}

func (e Endpoint) exampleRequestBody() string {
	val := ""
	if len(e.RequestExamples) > 0 {
		for _, example := range e.RequestExamples {
			if example.Key == "curl" {
				valBytes, err := json.MarshalIndent(example.RawValue, "		", "	")
				if err != nil {
					log.Fatal(err)
				}
				val = string(valBytes)
			}
		}
	}
	return val
}
