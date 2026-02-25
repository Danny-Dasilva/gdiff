.PHONY: build demo test

build:
	go build -o gdiff ./cmd/gdiff

demo: build
	./scripts/record-demos.sh

test:
	go test ./... -count=1
