# Swagger Parser

This repo contains a small utility to parse Swagger files and generate output `.md` files based on a Go template.

## Installation:

```bash
go install github.com/ubio/swagger-parser@latest
```

## Arguments:

It can take several arguments as flags:

| Flag | Example | Description
| --- | --- | ---
| `name` | `--name api` | The name of the API
| `template` | `--template template.gohtml` | The Go template file to use
| `pages` | `--pages ./path/to/pages.yaml` | The pages file to use
| `schema` | `--schema ./path/to/schema.yaml` | The schema file to use
| `output` | `--output ./path/to/output.md` | The output file to write

```bash
swagger-parser
    --name api
    --template template.gohtml
    --pages ./path/to/pages.yaml
    --schema ./path/to/schema.yaml
    --output ./path/to/output.md
```

