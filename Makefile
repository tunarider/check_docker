SHELL := /bin/bash
DEST := /usr/local/bin

BINARY=check_docker
PLATFORMS=darwin linux
ARCHITECTURES=amd64

build:
	mkdir -p bin/
	$(foreach GOOS, $(PLATFORMS), \
	          $(foreach GOARCH, $(ARCHITECTURES), \
                        $(shell export GOOS=$(GOOS); \
                                export GOARCH=$(GOARCH); \
                                cd cmd/$(BINARY) && go build -v -o ../../bin/$(BINARY)-$(GOOS)-$(GOARCH)/$(BINARY); \
                                cd ../../bin/ && tar czvf $(BINARY)-$(GOOS)-$(GOARCH).tar.gz $(BINARY)-$(GOOS)-$(GOARCH))))

install:
	cd cmd/$(BINARY) && go install
