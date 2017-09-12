VERSION     ?= 0.0.2
BINARIES    := linux/amd64/ecso windows/amd64/ecso.exe darwin/amd64/ecso
GITHUB_USER := bernos
GITHUB_REPO := ecso
RELEASE_DIR := bin/release
UPLOAD_LIST := $(foreach file, $(BINARIES), $(file)_upload)

NO_COLOR    := \033[0m
OK_COLOR    := \033[32;01m
ERROR_COLOR := \033[31;01m
WARN_COLOR  := \033[33;01m

all: $(BINARIES)

build:
	go build -o bin/local/ecso

clean:
	@echo "\n$(OK_COLOR)====> Cleaning$(NO_COLOR)"
	go clean ./... && rm -rf ./$(RELEASE_DIR)

clean-docs:
	rm -f ./docs/docs.md

deps:
	@echo "\n$(OK_COLOR)====> Fetching depenencies$(NO_COLOR)"
	go get -v github.com/aktau/github-release/...
	go get -v github.com/jteeuwen/go-bindata/...
	ls $(GOPATH)
	ls $(GOPATH)/bin

docs: test build clean-docs
	go run cmd/make-docs/main.go > ./docs/docs.md

install: test
	@echo "\n$(OK_COLOR)====> Installing$(NO_COLOR)"
	go install -ldflags "-X main.version=$(VERSION)"

release: tag $(BINARIES)
	@echo "\n$(OK_COLOR)====> Releasing v$(VERSION)$(NO_COLOR)"
	$(MAKE) .release GIT_TAG=$(shell git describe --abbrev=0 --tags)

tag:
	@echo "\n$(OK_COLOR)====> Tagging v$(VERSION)$(NO_COLOR)"
	git tag -a v$(VERSION) -m 'release $(VERSION)'

generate: deps
	@echo "\n$(OK_COLOR)====> Embedding assets$(NO_COLOR)"
	go generate -x ./...

test: generate
	@echo "\n$(OK_COLOR)====> Running tests$(NO_COLOR)"
	go test ./pkg/... ./cmd/...

.release: .create-github-release $(UPLOAD_LIST)

.create-github-release:
	git push && git push --tags
	github-release release \
		-u $(GITHUB_USER) \
		-r $(GITHUB_REPO) \
		-t $(GIT_TAG) \
		-n $(GIT_TAG)

$(UPLOAD_LIST): %_upload:
	github-release upload \
		-u $(GITHUB_USER) \
		-r $(GITHUB_REPO) \
		-t $(GIT_TAG) \
		-n $(subst /,-,$*) \
		-f $(RELEASE_DIR)/$*

$(BINARIES): osarch=$(subst /, ,$@)
$(BINARIES): test
	@echo "\n$(OK_COLOR)====> Building $@$(NO_COLOR)"
	GOOS=$(word 1, $(osarch)) GOARCH=$(word 2, $(osarch)) go build -a -ldflags "-X main.version=$(VERSION)" -o $(RELEASE_DIR)/$@ main.go

.PHONY: release tag test deps clean install docs generate
