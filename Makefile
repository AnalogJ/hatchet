.ONESHELL: # Applies to every targets in the file! .ONESHELL instructs make to invoke a single instance of the shell and provide it with the entire recipe, regardless of how many lines it contains.

########################################################################################################################
# Global Env Settings
########################################################################################################################
BINARY_NAME = hatchet
LD_FLAGS =
ifdef STATIC
LD_FLAGS := $(LD_FLAGS) -extldflags=-static
endif
ifdef GOOS
BINARY_NAME := $(BINARY_NAME)-$(GOOS)
LD_FLAGS := $(LD_FLAGS) -X main.goos=$(GOOS)
endif
ifdef GOARCH
BINARY_NAME := $(BINARY_NAME)-$(GOARCH)
LD_FLAGS := $(LD_FLAGS) -X main.goarch=$(GOARCH)
endif
ifdef GOARM
BINARY_NAME := $(BINARY_NAME)-$(GOARM)
endif
ifeq ($(OS),Windows_NT)
BINARY_NAME := $(BINARY_NAME).exe
endif

########################################################################################################################
# Binary
########################################################################################################################
.PHONY: all
all: build

.PHONY: clean
clean:
	go clean

.PHONY: generate
generate: dep
	go generate ./...

.PHONY: dep
dep:
	go mod vendor

.PHONY: test
test: dep
	go test -v -tags "static" ./...

.PHONY: test-coverage
test-coverage: dep
	go test -race -coverprofile=coverage.txt -covermode=atomic -v -tags "static" ./...

.PHONY: build
build: dep
	go build -ldflags "$(LD_FLAGS)" -o $(BINARY_NAME) -tags "static" ./cmd/hatchet
ifneq ($(OS),Windows_NT)
	chmod +x $(BINARY_NAME)
	file $(BINARY_NAME) || true
	ldd $(BINARY_NAME) || true
	./$(BINARY_NAME) || true
endif

########################################################################################################################
# Docker
########################################################################################################################