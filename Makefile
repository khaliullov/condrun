GOPATH := $(shell go env GOPATH)
GODEP  := $(GOPATH)/bin/dep
GOLINT := $(GOPATH)/bin/golint
BINARY_NAME := condrun
packages = $$(go list ./... | egrep -v '/vendor/')
files = $$(find . -name '*.go' | egrep -v '/vendor/')

.PHONY: all help run vet lint build install

all: build

help:           ## Show this help
	@echo "Usage:"
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

install:        ## Install binary
	cp $(BINARY_NAME) /usr/bin/

build:          ## Build the binary
build: vendor lint vet
	go build -o $(BINARY_NAME) cmd/main.go

run:            ## run script with arguments. example: `make run -- arg1 arg2`
run: vendor
	go run cmd/main.go $(filter-out $@, $(MAKECMDGOALS))

vet:            ## Run go vet
vet: vendor
	go vet -printfuncs=Debug,Debugf,Debugln,Info,Infof,Infoln,Error,Errorf,Errorln $(files)

lint:           ## Run go lint
lint: vendor $(GOLINT)
	$(GOLINT) -set_exit_status $(packages)

%:
	@true

$(GODEP):
	cd $(GOPATH) && go get -u github.com/golang/dep/cmd/dep

$(GOLINT):
	cd $(GOPATH) && go get -u golang.org/x/lint/golint
	cd $(GOPATH) && go get -u github.com/golang/lint/golint

Gopkg.toml: | $(GODEP)
	$(GODEP) init

vendor: | Gopkg.toml Gopkg.lock
	@echo "No vendor dir found. Fetching dependencies now..."
	GOPATH=$(GOPATH):. $(GODEP) ensure

