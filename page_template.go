package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"text/template"

	"github.com/ubio/swagger-parser/pkg/pages"

	"github.com/getkin/kin-openapi/openapi3"
)

type PageTemplate struct {
	Title       string
	Description string
	Endpoints   map[int]Endpoint
}

func parsePages(pages []pages.Page) {
	for _, page := range pages {
		template := NewPageTemplate()
		template.parsePage(page)
	}
}

func NewPageTemplate() PageTemplate {
	return PageTemplate{}
}

func (tpl *PageTemplate) parsePage(page pages.Page) {

	// sort the paths by the index key
	sort.Slice(page.Paths, func(i, j int) bool {
		return page.Paths[i].Index < page.Paths[j].Index
	})

	// build the template data
	pageTemplate := PageTemplate{
		Title:       page.Name,
		Description: page.Description,
		Endpoints:   make(map[int]Endpoint, 0),
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
				pathParams := make([]Param, 0)
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
						max := param.Value.Schema.Value.Max
						min := param.Value.Schema.Value.Min
						defaultVal := param.Value.Schema.Value.Default
						p := Param{
							Name:        param.Value.Name,
							Required:    param.Value.Required,
							Description: param.Value.Description,
							Type:        param.Value.Schema.Value.Type,
							Format:      param.Value.Schema.Value.Format,
							EnumJSON:    getEnumJSON(param.Value.Schema.Value.Enum),
							Example:     example.Value,
							MaxItems:    max,
							MinItems:    min,
							Default:     defaultVal,
						}
						params[param.Value.In] = append(params[param.Value.In], p)
						switch param.Value.In {
						case "query":
							queryParams = append(queryParams, p)
						case "header":
							headerParams = append(headerParams, p)
						case "path":
							pathParams = append(pathParams, p)
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
					PathParams:   pathParams,
				}
				endpoint.requestBody(matched)
				endpoint.requestBodyExamples(matched)
				endpoint.setServer(matched)
				endpoint.createExample("curl")
				endpoint.generateResponseExamples(matched)

				pageTemplate.Endpoints[p.Index] = *endpoint
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

func getEnumJSON(params []interface{}) string {
	escapedEnumJSONString := ""
	enum := make([]string, 0)
	for _, e := range params {
		enum = append(enum, fmt.Sprintf("%v", e))
	}
	if len(enum) == 0 {
		return escapedEnumJSONString
	}
	b, err := json.Marshal(enum)
	if err == nil {
		escapedEnumJSONString = template.HTMLEscaper(string(b))
	}
	return escapedEnumJSONString
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
			exampleJSONString := ""

			// check if this looks like json
			if isJSON(exampleString) {
				escaped := template.HTMLEscaper(exampleString)
				exampleJSONString = escaped
				exampleString = ""
			}

			paramType := param.Value.Type
			if paramType == "array" && param.Value.Items.Value.Type != "" {
				paramType = "array[" + param.Value.Items.Value.Type + "]"
			}

			p := RequestParam{
				Name:        k,
				Type:        paramType,
				Example:     exampleString,
				ExampleJSON: exampleJSONString,
				Description: param.Value.Description,
				Required:    required,
				EnumJSON:    getEnumJSON(param.Value.Enum),
				Default:     param.Value.Default,
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

		for i := 200; i < 400; i++ {
			if operation.Responses.Get(i) != nil {
				examples := operation.Responses.Get(i).Value.ExtensionProps.Extensions
				for _, example := range examples {
					e.ResponseExamples = append(e.ResponseExamples, e.createResponseExample(i, example)...)
				}
			}
		}

		for i := 400; i < 600; i++ {
			if operation.Responses.Get(i) != nil {
				examples := operation.Responses.Get(i).Value.ExtensionProps.Extensions
				for _, example := range examples {
					e.ResponseErrorExamples = append(e.ResponseErrorExamples, e.createResponseExample(i, example)...)
				}
			}
		}

		sort.Slice(e.ResponseExamples, func(i, j int) bool {
			return e.ResponseExamples[i].Key < e.ResponseExamples[j].Key
		})

		sort.Slice(e.ResponseErrorExamples, func(i, j int) bool {
			return e.ResponseErrorExamples[i].Key < e.ResponseErrorExamples[j].Key
		})

	}
}

func (e *Endpoint) createResponseExample(status int, example interface{}) []ResponseExample {

	exampleOutput := make([]ResponseExample, 0)

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
			Key:    k,
			Status: status,
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

		exampleOutput = append(exampleOutput, ex)
		i++
	}

	return exampleOutput
}
