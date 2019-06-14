clean::
	-rm -rf .build/*


# Buidling binaries

TRAVIS_COMMIT ?= dev.$(shell whoami).$(shell git rev-parse --short HEAD)
TRAVIS_REPO_SLUG ?= $(PACKAGE)
TRAVIS_JOB_NUMBER ?= $(shell whoami)

LDIMPORT=github.com/remerge/go-service
LDFLAGS=-X $(LDIMPORT).CodeVersion=$(TRAVIS_COMMIT) -X $(LDIMPORT).CodeBuild=$(TRAVIS_REPO_SLUG)\#$(TRAVIS_JOB_NUMBER)@$(shell date -u +%FT%TZ)

MAIN ?= main

# NOTE: logically pull out binaries from source tree
.build/$(PROJECT).%: $(GO_SOURCES) go.mod
	mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(basename $*) GOARCH=$(patsubst .%,%,$(suffix $*)) go build -o $@ -ldflags "$(LDFLAGS)" $(PACKAGE)/$(MAIN)

dist: .build/$(PROJECT).linux.amd64	## linux amd64 binary

.BIN_LOCAL = $(PROJECT).$(shell go env GOOS).$(shell go env GOARCH)

local: .build/$(.BIN_LOCAL)

# Deployment

release:
	git push origin master:production

deploy:
	/bin/bash -c 'cd ../chef.new && knife ssh roles:$(PROJECT) sudo chef-client'
