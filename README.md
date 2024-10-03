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
    sudo docker exec -it azflow-db bash
    root@03bbe2ac7448:/# psql -U postgres
    postgres=# create database azflowcore;
    create user azflow_admin with password 'abcd1234';
    alter role azflow_admin with superuser;
    alter database azflowcore owner to azflow_admin;
    # to list all dbs:
    \l
    # to list all users:
    \du

  - and put its credentials into a local copy of `.template.env` so the app can pickup and seed it. 
    - e.g: `docker run --name azflow-db -e POSTGRES_PASSWORD=abcd1234 -d -p 5432:5432 postgres`

To create secrets for a new k8s cluster env:
hung [ ~ ]$ nano .env -- paste all env fields here
hung [ ~ ]$ kubectl create secret generic azflow-api-env --from-env-file=.env

To enable SSL for FE:
```
sudo apt-get update
sudo apt-get install certbot python3-certbot-nginx
sudo certbot --nginx -d azflow.io
```

To login and pull docker container from azure container registry:
```
sudo docker login azflowcr.azurecr.io
sudo docker pull azflowcr.azurecr.io/azflowcr.azurecr.io/azflow-api:130
# run an interactive cli in the docker image
sudo docker run -it --entrypoint /bin/sh azflowcr.azurecr.io/azflowcr.azurecr.io/azflow-api:130
# copy file from running container
sudo docker cp 8c863a46712d:/root/azflow-api ./azflow-api
```


Useful debug commands:
```
azureuser@azflowmvp:~$ sudo docker run --rm --network azflow-network -it postgres psql "user=azflow-admin dbname=azflowcore password=abcd1234 host=azflow-db port=5432 sslmode=disable"
# to see what ip range and port the api is listening to:
azureuser@azflowmvp:~$ docker exec -it azflow-api /bin/sh
~ # netstat -tuln | grep 8080
#
docker restart azflow-api
docker stop azflow-api
docker rm azflow-api


```

for local dev specify ENV="local" and put the following into request header to authenticate:
```
{
  "Authorization": "azflow@local.dev"
}
```