#! /usr/bin/env bash

name=wgsltoy-postgres-dev

docker start "${name}"
if [ "$?" -eq "0" ]
then
  printf "\nExisting contiainer started!\n"
else
  printf "\nStarting new container!\n"
  mkdir -p tmp/data
  docker run -d --name="${name}" -v $(pwd)/tmp/data:/var/lib/postgresql/data -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=default -p 5432:5432 postgres:17
fi

printf "\nStarted on localhost:5432 (default port)\nPOSTGRES_USER=postgres\nPOSTGRES_PASSWORD=postgres\nPOSTGRES_DB=default\n"
