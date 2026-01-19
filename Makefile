build:
	@go build -v -o olympic cmd/olympic/main.go
run:
	@DB_FILE=olympic.db JWT_SECRET=testing-token-do-not-use-in-production-insecure-token go run -v cmd/olympic/main.go
clean:
	@go clean -v