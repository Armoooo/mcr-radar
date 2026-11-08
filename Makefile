.PHONY: run test check

run:
	go run ./cmd/mcr-radar

test:
	go test ./...

check:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')
	go vet ./...
	go test ./...
