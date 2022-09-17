info:
	@echo "Makefile is your friend"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

mock: ## generates mocks
	@go install github.com/vektra/mockery/v2@latest
	@mockery --name Pool --filename pool_mock.go --inpackage
	@mockery --name Tx --filename tx_mock.go --inpackage

test: ## runs test cases
	@- go test ./... -v > test.out
	@cat test.out