build:
	@CGO_ENABLED=1 go build -v -o cascade.bin cmd/cascade/main.go
run:
	@CGO_ENABLED=1 DB_FILE=cascade.db JWT_SECRET=testing-token-do-not-use-in-production-insecure-token go run -v cmd/cascade/main.go
clean:
	@go clean -v
test:
	@CGO_ENABLED=1 go test -v
build-container:
	@docker build -t cascade-app .