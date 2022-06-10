BINARY := $(shell go list -m)
SRC := $(shell ls *.go)
BUILD ?= .

.PHONY: clean

$(BINARY): go.mod $(SRC)
	go mod tidy
	go build -o $(BUILD)/$(BINARY)

clean:
	rm -rf $(BUILD)/$(BINARY)
