.PHONY: test test-race cover lint vet

test:
	go test -count=1 ./...

test-race:
	go test -race -count=1 ./...

cover:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out | tail -1
	@echo ""
	@echo "Per-package coverage:"
	@go tool cover -func=coverage.out | grep -E "^github.com" | awk '{print $$1, $$3}'

lint:
	golangci-lint run ./...

vet:
	go vet ./...