
DEFAULT: build

build:
	go mod tidy
	go build ./...
