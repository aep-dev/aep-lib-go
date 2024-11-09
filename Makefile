.PHONY: build
build:
	go build -o bin/myproject cmd/myproject/main.go

.PHONY: test
test:
	go test ./...

.PHONY: run
run:
	go run cmd/myproject/main.go

.PHONY: clean
clean:
	rm -rf bin/