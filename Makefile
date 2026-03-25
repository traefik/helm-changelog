VERSION="v0.0.3"

PROJECT_NAME="helm-changelog"
BINDIR ?= $(CURDIR)/bin
TMPDIR ?= $(CURDIR)/tmp
ARCH   ?= amd64

.PHONY: all $(MAKECMDGOALS)

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

test: ## Run unit-tests with coverage
	go test -cover ./...

build: ## Build binary
	mkdir -p $(BINDIR)
	CGO_ENABLED=0 go build -o ./bin/${PROJECT_NAME} .

verify: test build ## tests and builds

image: ## build docker image
	docker build -t ghcr.io/mloiseleur/${PROJECT_NAME}:${VERSION} .

clean: ## clean up created files
	rm -rf \
		$(BINDIR) \
		$(TMPDIR)

all: test build image ## Runs test, build and docker

lint: ## Run golint
	golint ./...

update-docs: build ## Upgrade automatic documentations
	bash hack/update-readme.sh
