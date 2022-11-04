#!/usr/bin/env bash
docker-compose up -d

# Deploy Interlink Database Schemas on a PostgreSQL container

echo "configuring DbMan for managing the interlink database"
dbman config use -n interlink
#dbman config set Repo.URI .
# for online use can set the URI to the http location of this project as shown below
dbman config set Repo.URI https://raw.githubusercontent.com/southwinds-io/interlink-db/master
dbman config set AppVersion 1.0.0
dbman config set Db.Name interlink
dbman config set Db.Host localhost
dbman config set Db.Port 5432
dbman config set Db.Username interlink
dbman config set Db.Password 1nt3rl1nk
dbman config set Db.AdminUsername postgres
dbman config set Db.AdminPassword p0stgr3s

echo "waiting for database server to start"
sleep 2

echo "creating database"
dbman db create

echo "deploying database schemas"
dbman db deploy

echo "configuring DbMan for managing the pilotctl database"
dbman config use -n pilotctl
#dbman config set Repo.URI .
# for online use can set the URI to the http location of this project as shown below
dbman config set Repo.URI https://raw.githubusercontent.com/southwinds-io/pilotctl-db/master
dbman config set AppVersion 1.0.0
dbman config set Db.Name pilotctl
dbman config set Db.Host localhost
dbman config set Db.Port 5432
dbman config set Db.Username pilotctl
dbman config set Db.Password p1l0tctl
dbman config set Db.AdminUsername postgres
dbman config set Db.AdminPassword p0stgr3s

echo "waiting for database server to start"
sleep 2

echo "creating database"
dbman db create

echo "deploying database schemas"
dbman db deploy

