.PHONY: lint

lint:
	cd BackEnd && go mod tidy && go mod download
	golangci-lint run ./BackEnd/...