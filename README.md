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

## Pages File

The `pages` file is an opinionated YAML file that contains the configuration f which pages to generate. For example:

```yaml
// pets.yaml

pages:
  - name: Pets
    filename: pets.md
    description: |
      Pets are excellent friends. Why not get one via the API.
    paths:
      - method: get
        path: /pets
      - method: post
        path: /pets
```

Will generate a page and include the `GET /pets` and `POST /pets` spec in the output passed to the go template file, where the `name` and `description` will also be printed.

## Example

```bash
go run *.go \
    --name pets \
    --template ./example/template.gohtml \
    --pages ./example/pets.yaml \
    --schema ./example/schema.yaml \
    --output ./example/output
```
