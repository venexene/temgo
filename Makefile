.PHONY: build test clean install vet

build:
	go build -o temgo .

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

clean:
	rm -f temgo

install:
	go install .