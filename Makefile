VERSION     ?= 0.0.1
BINARIES    := linux/amd64/ecso windows/amd64/ecso.exe darwin/amd64/ecso
GITHUB_USER := bernos
GITHUB_REPO := ecso
RELEASE_DIR := bin/release
UPLOAD_LIST := $(foreach file, $(BINARIES), $(file)_upload)

NO_COLOR    := \033[0m
OK_COLOR    := \033[32;01m
ERROR_COLOR := \033[31;01m
WARN_COLOR  := \033[33;01m

DOC_COMMANDS := init \
				environment_add \
				environment_up \
				environment_describe \
				environment_down \
				environment_rm \
				env \
				service_add \
				service_up \
				service_down \
				service_ls \
				service_ps \
				service_logs \
				service_describe

all: $(BINARIES)

build:
	go build -o bin/local/ecso

clean:
	@echo "\n$(OK_COLOR)====> Cleaning$(NO_COLOR)"
	go clean ./... && rm -rf ./$(RELEASE_DIR)

clean-docs:
	rm -f ./docs.md

deps:
	@echo "\n$(OK_COLOR)====> Fetching depenencies$(NO_COLOR)"
	go get github.com/aktau/github-release

# docs: test build clean-docs $(DOC_COMMANDS)

docs: test build clean-docs
	go run cmd/make-docs/main.go > docs.md

install: test
	@echo "\n$(OK_COLOR)====> Installing$(NO_COLOR)"
	go install -ldflags "-X main.version=$(VERSION)"

release: tag $(BINARIES)
	@echo "\n$(OK_COLOR)====> Releasing v$(VERSION)$(NO_COLOR)"
	$(MAKE) .release GIT_TAG=$(shell git describe --abbrev=0 --tags)

tag:
	@echo "\n$(OK_COLOR)====> Tagging v$(VERSION)$(NO_COLOR)"
	git tag -a v$(VERSION) -m 'release $(VERSION)'

test: deps
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

$(DOC_COMMANDS):
	@echo "# ecso $(subst _, ,$@)\n\n\`\`\`" >> docs.md
	@./bin/local/ecso $(subst _, ,$@) -h >> docs.md
	@echo "\`\`\`\n" >> docs.md

.PHONY: release tag test deps clean install
