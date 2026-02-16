# disable default rules
.SUFFIXES:
MAKEFLAGS+=-r -R
DATE  = $(shell date +%Y%m%d%H%M%S)
export GOBIN = $(CURDIR)/bin

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
