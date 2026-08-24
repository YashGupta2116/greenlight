## Bin Directory

- It will contain our compiled application binaries

## cmd/api

- it will contain the application specific code for our greenlight api application
- This will include the code for running the server, reading and writing HTTP
  requests, and managing authentication

## internal

- it will contain various ancillary packages used by our API.
- It will contain the code for interacting with our database, doing data validation, sending emails
  and so on. Basically, any code which isn’t application-specific and can potentially be
  reused will live in here.
- Our Go code under cmd/api will import the packages in the internal directory (but never the other way around).

## migrations

- this directory will contain the SQL migration files for our database.

## remote

- this directory will contain the configuration files and setup scripts for our
  production server.

## go.mod

- this file will declare our project dependencies, versions and module path.

## Makefile

- this file will contain recipes for automating common administrative tasks — like
  auditing our Go code, building binaries, and executing database migrations.
