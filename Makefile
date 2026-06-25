.PHONY: build test clean install

build:
	go build -o temgo ./cmd/temgo
	go build -o temgo-tui ./cmd/temgo-tui

test:
	go test -race ./...

clean:
	rm -f temgo temgo-tui

install:
	go install ./cmd/temgo
	go install ./cmd/temgo-tui