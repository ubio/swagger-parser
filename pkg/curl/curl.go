package curl

import (
	"fmt"
	"strings"
)

type Command struct {
	Server        string
	Method        string
	Path          string
	Headers       []string
	QueryParams   []string
	Body          string
	ExampleString string
}

func NewCommand(server string, method string, path string, headers []string, queryParams []string, body string) *Command {
	c := &Command{
		Method:      method,
		Server:      server,
		Path:        path,
		Headers:     headers,
		QueryParams: queryParams,
		Body:        body,
	}
	c.GenerateExample()
	return c
}

func (c *Command) GenerateExample() {
	c.ExampleString = fmt.Sprintf(`curl -X %s '%s%s'`, strings.ToUpper(c.Method), c.Server, c.Path)
	c.addParams("headers")
	c.addParams("query")
	c.addBody()
}

func (c *Command) addParams(paramType string) {

	params := []string{}
	curlFlag := ""

	switch paramType {
	case "headers":
		params = c.Headers
		curlFlag = "-H"
	case "query":
		params = c.QueryParams
		curlFlag = "-d"
	}

	if len(params) > 0 {

		c.ExampleString += ` \`

		if paramType == "query" {
			c.ExampleString += `
		-G \`
		}

		for i, param := range params {
			if param != "" {
				c.ExampleString += fmt.Sprintf(`
		%s '%s'`, curlFlag, param)
				if i != len(params)-1 {
					c.ExampleString += ` \`
					continue
				}
			}
		}
	}
}

func (c *Command) addBody() {
	if strings.TrimSpace(c.Body) != "" {
		c.ExampleString += fmt.Sprintf(` \
		-d@- <<EOF
		%s
		EOF`, strings.TrimSpace(c.Body))
	}
}
