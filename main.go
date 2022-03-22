package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	err              error
	swagger          *openapi3.T
	endpointTemplate *template.Template

	// set by flags:
	apiName    string
	goTemplate string
	pagesFile  string
	schemaFile string
	outputDir  string
)

func init() {

	parseFlags()

	// check the schema files supplied exist
	if err := verifyFilesExist(); err != nil {
		log.Fatal("verifySchemasExist() failed: ", err)
	}

	// load the swagger file
	swagger, err = openapi3.NewLoader().LoadFromFile(schemaFile)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	endpointTemplate, err = template.ParseFiles(goTemplate)
	if err != nil {
		log.Fatal(err)
	}
	pages := getPages()
	parsePages(pages)
}

func parseFlags() {
	apiNameFlag := flag.String("name", "", "The name of the API")                  // the name of the API
	templateFileFlag := flag.String("template", "", "The Go template file to use") // the go template
	pagesFlag := flag.String("pages", "", "The pages file to use")                 // the pages yaml file
	schemaFlag := flag.String("schema", "", "The schema file to use")              // the swagger yaml file
	outputFlag := flag.String("output", "", "The output file to write")            // the output directory
	flag.Parse()

	apiName = *apiNameFlag
	goTemplate = *templateFileFlag
	pagesFile = *pagesFlag
	schemaFile = *schemaFlag
	outputDir = *outputFlag
	outputDir = fmt.Sprintf("%s/%s/", outputDir, apiName)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.MkdirAll(outputDir, 0755)
	}
}

func verifyFilesExist() error {
	// check the pages file exists
	if _, err := os.Stat(pagesFile); err != nil {
		return fmt.Errorf("Pages file does not exist: %s", pagesFile)
	}

	// check the schema exists
	if _, err := os.Stat(schemaFile); err != nil {
		return fmt.Errorf("Schema file does not exist: %s", schemaFile)
	}

	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
