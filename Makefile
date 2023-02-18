
test_full: ## Run all tests, including integration tests with database.
	DATABASE_URL="postgres://postgres:postgres@localhost:54323/postgres" go test ./...
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

run_server:
	DATABASE_DSN="postgres://postgres:postgres@localhost:54323/postgres" go run cmd/server/main.go -i 1s
