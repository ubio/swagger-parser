package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

type PageTemplate struct {
	Title       string
	Description string
	Endpoints   []Endpoint
}

func parsePages(pages []Page) {
	for _, page := range pages {
		template := NewPageTemplate()
		template.parsePage(page)
	}
}

func NewPageTemplate() PageTemplate {
	return PageTemplate{}
}

func (tpl *PageTemplate) parsePage(page Page) {

	// build the template data
	pageTemplate := PageTemplate{
		Title:       page.Name,
		Description: page.Description,
		Endpoints:   make([]Endpoint, 0),
	}

	for path, pathInfo := range swagger.Paths {
		for _, p := range page.Paths {

			if p.Path == path {

				var matched *openapi3.Operation

				switch p.Method {
				case "post":
					matched = pathInfo.Post
				case "get":
					matched = pathInfo.Get
				case "put":
					matched = pathInfo.Put
				case "options":
					matched = pathInfo.Options
				}

				queryParams := make([]Param, 0)
				headerParams := make([]Param, 0)
				params := make(map[string][]Param)
				if matched != nil {
					for _, param := range matched.Parameters {

						exampleBody, err := json.Marshal(param.Value.Example)
						if err != nil {
							log.Fatal(err)
						}
						example := Example{}
						if err := json.Unmarshal(exampleBody, &example); err != nil {
							log.Fatal(err)
						}
						p := Param{
							Name:        param.Value.Name,
							Required:    param.Value.Required,
							Description: param.Value.Description,
							Type:        param.Value.Schema.Value.Type,
							Example:     example.Value,
						}
						params[param.Value.In] = append(params[param.Value.In], p)
						switch param.Value.In {
						case "query":
							queryParams = append(queryParams, p)
						case "header":
							headerParams = append(headerParams, p)
						}
					}
				}

				endpoint := &Endpoint{
					Path:         path,
					Method:       p.Method,
					Info:         matched,
					Params:       params,
					QueryParams:  queryParams,
					HeaderParams: headerParams,
				}
				endpoint.requestBody(matched)
				endpoint.requestBodyExamples(matched)
				endpoint.setServer(matched)
				endpoint.createExample("curl")
				endpoint.generateResponseExamples(matched)

				pageTemplate.Endpoints = append(pageTemplate.Endpoints, *endpoint)
			}
		}
	}

	f, err := os.Create(fmt.Sprintf("%s%s", outputDir, page.Filename))
	if err != nil {
		log.Fatal(err)
	}

	err = endpointTemplate.Execute(f, pageTemplate)
	if err != nil {
		log.Fatal(err)
	}
}

func isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func (e *Endpoint) requestBody(operation *openapi3.Operation) {
	if operation.RequestBody != nil {
		properties := operation.RequestBody.Value.Content["application/json"].Schema.Value.Properties
		requiredFields := operation.RequestBody.Value.Content["application/json"].Schema.Value.Required
		for k, param := range properties {
			var required bool
			if contains(requiredFields, k) {
				required = true
			}

			exampleString := string(param.Value.Example.(string))
			exampleObjectString := ""

			// check if this looks like a json object
			if isJSON(exampleString) {
				escaped := template.HTMLEscaper(exampleString)
				fmt.Println(escaped)
				exampleObjectString = escaped
				exampleString = ""
			}

			p := RequestParam{
				Name:          k,
				Type:          param.Value.Type,
				Example:       exampleString,
				ExampleObject: exampleObjectString,
				Description:   param.Value.Description,
				Required:      required,
			}
			e.RequestParams = append(e.RequestParams, p)
		}
	}
}

type RequestExamples []RequestExample

type RequestExample struct {
	Key         string
	Value       string
	RawValue    interface{}
	Summary     string
	Description string
	Required    bool
}

func (e *Endpoint) requestBodyExamples(operation *openapi3.Operation) {
	e.RequestExamples = make([]RequestExample, 0)
	if operation.RequestBody != nil {
		specExamples := operation.RequestBody.Value.Content.Get("application/json").Examples
		for key, specExample := range specExamples {
			example := RequestExample{}
			example.Key = key
			example.Description = specExample.Value.Description
			example.Summary = specExample.Value.Summary
			example.RawValue = specExample.Value.Value
			val, err := json.MarshalIndent(specExample.Value.Value, "", "    ")
			if err != nil {
				log.Fatal(err)
			}
			example.Value = string(val)
			e.RequestExamples = append(e.RequestExamples, example)
		}
	}

}

func (e *Endpoint) setServer(operation *openapi3.Operation) {
	server := ""
	for _, svr := range swagger.Servers {
		server = svr.URL
	}

	if operation.Servers != nil {
		servers, err := json.Marshal(operation.Servers)
		if err != nil {
			log.Fatal(err)
		}
		ss := []Server{}
		err = json.Unmarshal(servers, &ss)
		if err != nil {
			log.Fatal(err)
		}
		if len(ss) > 0 {
			for _, s := range ss {
				server = s.URL
			}
		}
	}
	e.Server = server
}

func (e *Endpoint) generateResponseExamples(operation *openapi3.Operation) {
	e.ResponseExamples = make([]ResponseExample, 0)
	if operation.Responses != nil {
		examples := operation.Responses.Get(200).Value.ExtensionProps.Extensions
		for _, example := range examples {

			mp := make(map[string]interface{})
			err := json.Unmarshal([]byte(example.(json.RawMessage)), &mp)
			if err != nil {
				log.Fatal(err)
			}
			i := 0
			for k, v := range mp {

				e.ResponseExampleKeys = e.ResponseExampleKeys + k
				if i != len(mp)-1 {
					e.ResponseExampleKeys = e.ResponseExampleKeys + ","
				}

				ex := ResponseExample{
					Key: k,
				}

				for k2, v2 := range v.(map[string]interface{}) {
					switch k2 {
					case "summary":
						ex.Summary = v2.(string)
					case "value":
						exampleBytes, err := json.MarshalIndent(v2, "", "    ")
						if err != nil {
							log.Fatal(err)
						}
						ex.Value = string(exampleBytes)
					}
				}

				e.ResponseExamples = append(e.ResponseExamples, ex)
				i++
			}

			sort.Slice(e.ResponseExamples, func(i, j int) bool {
				return e.ResponseExamples[i].Key < e.ResponseExamples[j].Key
			})
		}
	}
}
