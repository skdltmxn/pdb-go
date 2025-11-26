.PHONY: build clean test vet

BINARY := pdbview
BIN_DIR := bin

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) ./cmd/pdbview

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)
