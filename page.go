package parser

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

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

				queryParamBytes, err := json.Marshal(queryParams)
				if err != nil {
					log.Fatal(err)
				}
				headerParamBytes, err := json.Marshal(headerParams)
				if err != nil {
					log.Fatal(err)
				}

				endpoint := &Endpoint{
					Path:         path,
					Method:       p.Method,
					Info:         matched,
					Params:       params,
					QueryParams:  sanitizeData(string(queryParamBytes)),
					HeaderParams: sanitizeData(string(headerParamBytes)),
				}
				endpoint.requestBody(matched)
				endpoint.requestBodyExamples(matched)
				endpoint.setServer(matched)
				endpoint.curlExample(matched)
				endpoint.generateResponseExamples(matched)

				pageTemplate.Endpoints = append(pageTemplate.Endpoints, *endpoint)
			}
		}
	}

	f, err := os.Create(fmt.Sprintf("%s%s%s", outputDir, baseDir, page.Filename))
	if err != nil {
		log.Fatal(err)
	}

	err = endpointTemplate.Execute(f, pageTemplate)
	if err != nil {
		log.Fatal(err)
	}
}
