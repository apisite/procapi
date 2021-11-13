# pgcall package Makefile

SHELL         = /bin/bash
CFG           = .env
GO           ?= go
GOSOURCES    ?= ./...
#./ginproc

CODECOV_KEY   =

# Random id for test objects names
RANDOM_ID       ?= $(shell < /dev/urandom tr -dc A-Za-z0-9 | head -c14; echo)

# Postgresql Database image
PG_IMAGE     ?= postgres:11.2

PRG          ?= $(shell basename $$PWD)

# Postgresql variables
PGDATABASE   ?= $(PRG)
PGUSER       ?= $(PRG)
PGAPPNAME    ?= $(PRG)
PGHOST       ?= localhost
PGPORT       ?= 5431
PGSSLMODE    ?= disable
PGPASSWORD   ?= $(shell < /dev/urandom tr -dc A-Za-z0-9 | head -c14; echo)

# Running postgresql container name for `docker exec`
DB_CONTAINER ?= procapi_$(RANDOM_ID)
#dcape_db_1

define CONFIG_DEFAULT
# ------------------------------------------------------------------------------
# pgcall config file, generated by make $(CFG)

# Database

# Host
PGHOST=$(PGHOST)
# Port
PGPORT=$(PGPORT)
# Name
PGDATABASE=$(PGDATABASE)
# User
PGUSER=$(PGUSER)
# Password
PGPASSWORD=$(PGPASSWORD)

# docker postgresql container name
DB_CONTAINER=$(DB_CONTAINER)


# codecov.io API key
CODECOV_KEY=$(CODECOV_KEY)

endef
export CONFIG_DEFAULT

# ------------------------------------------------------------------------------

-include $(CFG)
export

.PHONY: gen config

# ------------------------------------------------------------------------------

## update generated mocks
# by https://github.com/golang/mock/
gen:
	$(GO) generate

## run linter
lint:
	golangci-lint run $(GOSOURCES)

# ------------------------------------------------------------------------------

## Show coverage
cov:
	TEST_UPDATE=yes $(GO) test -coverprofile=coverage.txt -race -covermode=atomic -v $(GOSOURCES)

## Show coverage
cov-db:
	SCHEMA="rpc,public" TZ="Europe/Berlin" \
	$(GO) test -coverprofile=coverage.out -race -covermode=atomic -tags=db -v $(GOSOURCES)

cov-db-upd:
	SCHEMA="rpc,public" TEST_UPDATE=yes \
	$(GO) test -coverprofile=coverage.out -race -covermode=atomic -tags=db -v $(GOSOURCES)

## Show package coverage in html
cov-html:
	$(GO) tool cover -html=coverage.out

cov-cmp:
	$(MAKE) -s cov-db
	@sort < coverage.out > coverage-db.out
	$(MAKE) -s cov
	@sort < coverage.out > coverage-mock.out
	@diff -c0 coverage-mock.out coverage-db.out > coverage.diff &&  echo "No differences" || less coverage.diff

## Format go sources
fmt:
	$(GO) fmt ./lib/... && $(GO) fmt ./counter/... && $(GO) fmt ./cmd/...

## Run vet
vet:
	$(GO) vet -tags db *.go
	$(GO) vet pgtype/*.go
	$(GO) vet ginproc/*.go

# ------------------------------------------------------------------------------

# Run tests when postgresql is available
test-db-exists:

# ------------------------------------------------------------------------------
# Run tests with docker

# find unused local port
# https://unix.stackexchange.com/questions/55913/whats-the-easiest-way-to-find-an-unused-local-port
# https://unix.stackexchange.com/a/248319
find-port:
	@if [[ ! "$(PGPORT)" ]] ; then  \
	  read LOWERPORT UPPERPORT < /proc/sys/net/ipv4/ip_local_port_range ; \
	  while true ; do  \
	    PGPORT="`shuf -i $$LOWERPORT-$$UPPERPORT -n 1`" ; \
	    ss -lpn | grep -q ":$$PGPORT " || break ; \
	  done ; \
	fi
	echo $(PGPORT)

# Start postgresql via docker
test-docker-run:
	@docker run --rm --name $$DB_CONTAINER \
	-p "127.0.0.1:$$PGPORT:5432" \
	-e POSTGRES_PASSWORD=$$PGPASSWORD \
	-e POSTGRES_DB=$$PGDATABASE \
	-e WORKDIR=/docker-entrypoint-initdb.d \
	-v $(shell pwd)/tmp-db:/var/lib/postgresql/data \
	-v $(shell pwd)/testdata:/docker-entrypoint-initdb.d $$PG_IMAGE

# TODO: ALTER DATABASE db WITH ALLOW_CONNECTIONS false;

## Run psql via docker
psql-docker:
	@docker exec -ti $$DB_CONTAINER psql -U $$PGUSER -d $$PGDATABASE

## Run local psql
psql:
	@psql

# Stop postgresql via docker
test-docker-stop:
	docker stop $$DB_CONTAINER

# Count lines of code (including tests)
cloc:
	cloc --by-file --not-match-f='(_mock_test.go|.sql|ml|Makefile|resource.go)$$' .

# ------------------------------------------------------------------------------

## create initial config
$(CFG):
	@[ -f $@ ] || { echo "Creating default $@" ; echo "$$CONFIG_DEFAULT" > $@ ; }

## Create default $(CFG) file
config:
	@true

# ------------------------------------------------------------------------------

## List Makefile targets
help:
	@grep -A 1 "^##" Makefile | less

##
## Press 'q' for exit
##
