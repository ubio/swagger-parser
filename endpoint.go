package main

import (
	"github.com/getkin/kin-openapi/openapi3"
)

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

type RequestParams []RequestParam

type RequestParam struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Example     interface{} `json:"example"`
	Enum        []string    `json:"enum"`
	Required    bool        `json:"required"`
}
