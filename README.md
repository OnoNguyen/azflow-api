# azflow-api
Prerequisites:
- Go installed

Tech stacks:
- Golang:
  - To run the app, type `air` in the terminal.
- gaphql:
  - this project use genql to generate graphql schema:
    - run `go generate ./...` or `.make gen` to re-generate `schema.graphqls`, `schema.resolvers.go` and `generated.go`.
  - note: remember to run the above anytime there's a change in the resolvers or schema.
- postgres:
  - pull docker postgres `docker pull postgres` and start the instance
  - create an empty Postgres db (use pgAdmin or CLI) and put its credentials into a local copy of `.template.env` so the app can pickup and seed it.