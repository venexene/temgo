.PHONY: build test clean install

build:
	go build -o temgo ./cmd/cli
	go build -o temgo-tui ./cmd/tui

test:
	go test -race ./...

clean:
	rm -f temgo temgo-tui

install:
	go install ./cmd/cli
	go install ./cmd/tui