# slotify-backend

## Set up

1. Install `pre-commit`:

```bash
    pip install pre-commit
```

2. Install the hooks:

```bash
    pre-commit install
```

This will make sure that golang ci lint is ran before you commit so we don't need to
fix errors after pushing.

3. Make sure Go is installed
4. Run the app through:

```bash
    make run
```

(See other options in the Makefile)

## OpenAPI

Our OpenAPI spec can be found at openapi.yaml

[oapi-codegen Go lib](https://github.com/oapi-codegen/oapi-codegen) is used to generate server code
based on openapi.yaml, so things like input validation, registering handlers among other things are
automated.

/api/oapi_codegen_cfg.yaml defines our oapi-codegen config (eg. what Go server to use), to see everything that can be stated in this file see the oapi_codegen_cfg_schema.json

Using OpenAPI means API documentation (routes, parameters, etc.) is generated for us, which can be done by doing:

```bash
make generate_api_docs
```

This will regenerate the API docs which can be found under api_docs/
