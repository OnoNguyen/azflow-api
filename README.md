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
  - install docker: https://docs.docker.com/engine/install/
  - pull docker postgres `docker pull postgres` and start the instance
  - create an empty Postgres db using pgAdmin or CLI, e.g:
    ```
    sudo docker exec -it 03bbe2ac7448 bash
    root@03bbe2ac7448:/# psql -U postgres
    postgres=# create database azflowcore;

  - and put its credentials into a local copy of `.template.env` so the app can pickup and seed it. 
    - e.g: `docker run --name azflow-db -e POSTGRES_PASSWORD=abcd1234 -d -p 5432:5432 postgres`

To create secrets for a new k8s cluster env:
hung [ ~ ]$ nano .env -- paste all env fields here
hung [ ~ ]$ kubectl create secret generic azflow-api-env --from-env-file=.env

To enable SSL for FE:
```
sudo apt-get update
sudo apt-get install certbot python3-certbot-nginx
sudo certbot --nginx -d azflow.australiasoutheast.cloudapp.azure.com