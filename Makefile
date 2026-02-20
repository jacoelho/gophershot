# disable default rules
.SUFFIXES:
MAKEFLAGS+=-r -R
DATE  = $(shell date +%Y%m%d%H%M%S)
export GOBIN = $(CURDIR)/bin
EXAMPLE_DIR = example

$(GOBIN)/staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: gophershot
gophershot: $(GOBIN)/gophershot

$(GOBIN)/gophershot:
	go build -o $(GOBIN)/gophershot ./cmd/gophershot

.PHONY: staticcheck
staticcheck: $(GOBIN)/staticcheck
	$(GOBIN)/staticcheck ./...

.PHONY: test
test:
	go test -timeout 2m -race -shuffle=on ./...

.PHONY: example example-go example-tf
example: example-go example-tf

example-go: $(GOBIN)/gophershot
	$(GOBIN)/gophershot --out $(EXAMPLE_DIR)/go-example.png --lines 55:89 internal/app/run.go

example-tf: $(GOBIN)/gophershot
	$(GOBIN)/gophershot --out $(EXAMPLE_DIR)/tf-example.png $(EXAMPLE_DIR)/main.tf
