GOCMD=go
GOBUILD=$(GOCMD) build
GOMOD=$(GOCMD) mod
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

all: build

build:
	@echo "Building SMTP-NC...."
	$(GOBUILD) -v -ldflags="-extldflags=-static" -o "smtp-nc" main.go

test:
	@./smtp-nc
clean:
	@rm -rf ./smtp-nc
help:
	@echo make build
	@echo make test
	@echo make test
