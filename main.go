package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

// Always expect errors
var err error

// name is the name of the swagger and pages file to parse as an input
var name string

// the ref to swagger.T
var swagger *openapi3.T

// outputDir tells us where to write files
var outputDir string

const (
	// API_REF is the template to use for the docs
	API_REF = "./cmd/parse-swagger/api.gohtml"
)

// these vars are based on $name and point to the relevant pages exist
var (
	schemaFile       string
	baseDir          string
	pagesFile        string
	endpointTemplate *template.Template
)

// init configures the command for the correct swagger file and output
// based on the flags provided. For best practice, add your API Ref to
// `/schemas/{name}/schema.yaml` and for your additional content pages
// to `/schemas/{name}/pages.yaml` and pass in the flags:
//
//		-name	$NAME
//
// 		-base	./src/$NAME/
//		-pages	schemas/$NAME/pages.yaml
//		-ref	schemas/$NAME/schema.yaml
//
func init() {

	// get the name of the API
	nameFlag := flag.String("name", "", "")
	flag.Parse()

	name = *nameFlag

	// set the base directory for this API from the name
	baseDir := fmt.Sprintf("schemas/%s", name)

	// set the schema file from the name
	schemaFile = fmt.Sprintf("%s/schema.yaml", baseDir)

	// set the pages file from the name
	pagesFile = fmt.Sprintf("%s/pages.yaml", baseDir)

	// check the schema exists from the name
	if _, err := os.Stat(schemaFile); err != nil {
		log.Fatal("Schema doesn't exist", err, schemaFile)
	}

	// check the pages exist from the name
	if _, err := os.Stat(pagesFile); err != nil {
		log.Fatal("Pages don't exist", err, pagesFile)
	}

	// load the swagger file through a loader
	swagger, err = openapi3.NewLoader().LoadFromFile(schemaFile)
	if err != nil {
		log.Fatal(err)
	}

	outputDir = fmt.Sprintf("./src/%s/", name)
}

type Param struct {
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Example     string `json:"example"`
}

type Example struct {
	Value string `json:"value"`
}

type Endpoint struct {
	Server                  string
	Path                    string
	Method                  string
	Info                    *openapi3.Operation
	Params                  map[string][]Param
	QueryParams             string
	HeaderParams            string
	Curl                    string
	RequestParams           RequestParams
	RequestParamsMarshalled string
	RequestExamples         RequestExamples
	ResponseExamples        ResponseExamples // this is rendered in a slot
	ResponseExampleKeys     string           // this is passed to Vue as a csv
}

type ResponseExamples []ResponseExample

type ResponseExample struct {
	Key     string
	Value   string
	Summary string
}

type Server struct {
	URL string `json:"url"`
}

func main() {

	endpointTemplate, err = template.ParseFiles(API_REF)
	if err != nil {
		log.Fatal(err)
	}
	pages := getPages()
	parsePages(pages)
}

type RequestParams []RequestParam

type RequestParam struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Example     interface{} `json:"example"`
	Enum        []string    `json:"enum"`
	Required    bool        `json:"required"`
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
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
			p := RequestParam{
				Name:        k,
				Type:        param.Value.Type,
				Example:     param.Value.Example,
				Description: param.Value.Description,
				Required:    required,
			}
			e.RequestParams = append(e.RequestParams, p)
		}
		marshalled, err := json.Marshal(e.RequestParams)
		if err != nil {
			log.Fatal(err)
		}
		e.RequestParamsMarshalled = sanitizeData(string(marshalled))
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

func (e *Endpoint) curlExample(operation *openapi3.Operation) {
	curl := fmt.Sprintf(`curl -X %s '%s%s' \`, e.Method, e.Server, e.Path)
	for i, param := range e.Params["header"] {
		if param.Example != "" {
			curl += fmt.Sprintf(`
	-H '%s'`, param.Example)
		}
		if i != len(e.Params["header"])-1 {
			curl += ` \`
			continue
		}
		if len(e.Params["query"]) > 0 {
			curl += ` \`
		}
	}
	if len(e.Params["query"]) > 0 {
		curl += `
	-G \`
	}
	for i, param := range e.Params["query"] {
		if param.Example != "" {
			curl += fmt.Sprintf(`
	-d '%s'`, param.Example)
		}
		if i != len(e.Params["query"])-1 {
			curl += ` \`
		}
	}
	if len(e.RequestExamples) > 0 {
		for _, example := range e.RequestExamples {
			if example.Key == "curl" {

				val, err := json.MarshalIndent(example.RawValue, "    ", "    ")
				if err != nil {
					log.Fatal(err)
				}
				curl += fmt.Sprintf(`
	-d@- <<EOF
	%s
EOF`, string(val))
			}
		}
	}

	e.Curl = curl
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

func sanitizeData(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}
