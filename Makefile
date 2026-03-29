.PHONY: run build test clean

BINARY := coa-server

run:
	go run .

build:
	go build -o $(BINARY) .

test:
	go test ./...

clean:
	go clean
	rm -f $(BINARY)
