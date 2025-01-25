# slotify-backend

## Set up

1. Clone repo, there is a submodule so you must use:

```bash
git clone --recurse-submodules
```

2. Install `make`

3. Install `pre-commit`:

```bash
    pip install pre-commit
```

4. Install the hooks:

```bash
    pre-commit install
```

This will make sure that golang ci lint is ran before you commit so we don't need to
fix errors after pushing.

5. Install Golang (Go is still needed for go generate ./..., if this is a problem we can dockerise)
6. Run the app through:

```bash
    make generate # locally runs go and generates files needed, must be run
    make run # runs docker compose up, starts up golang server and db in containers
    make stop # docker compose down, stops the above containers

```

(See other options in the Makefile)

## Project Structure

```
.
├── api # API code
├── api_docs # API Docs that were automatically generated (see below)
├── cmd # Entry point to our API
├── database # Database utilities
├── generate # Defines go files for oapi-codegen generation (see below)
├── go.mod
├── go.sum
├── integration # Integration tests
├── Makefile # Defines commands to use for this app (eg. make run)
├── oapi_codegen_cfg_schema.json # schema for oapi-codegen options (see below)
├── README.md
├── shared # shared repo submodule (containing docker containers and db schema)
├── testutil # test utilities
└── tools # Folder is needed for oapi-codegen dependencies, so it doesn't clutter our main go.mod
```

## OpenAPI

Our OpenAPI spec can be found at openapi.yaml

[oapi-codegen Go lib](https://github.com/oapi-codegen/oapi-codegen) is used to generate server code
based on `shared/openapi/openapi.yaml`, so things like input validation,
registering handlers among other things are automated.

`/generate/oapi_codegen_cfg.yaml` defines our oapi-codegen config (eg. what Go server to use), to see everything that can be stated in this file see the oapi_codegen_cfg_schema.json

Using OpenAPI means API documentation (routes, parameters, etc.) is generated for us, which can be done by doing:

```bash
make generate_api_docs
```

This will regenerate the API docs which can be found under api_docs/

```bash
make generate
```

This will generate API code based on our spec into a file named `api/server.gen.go`
