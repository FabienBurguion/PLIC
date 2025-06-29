.PHONY: all
all: build

.PHONY: build
build:
	cd http-handler && go build -o ../bin/http-handler.exe

.PHONY: docker-up
docker-up:
	docker compose up -d

.PHONY: docker-down
docker-down:
	docker compose down -v

.PHONY: test
test:
	docker compose up -d
	go test ./... -v -coverprofile=coverage.out
	docker compose down -v

.PHONY: zip-windows
zip-windows:
	powershell -ExecutionPolicy Bypass -File zip-project.ps1

.PHONY: deploy-windows
deploy-windows: zip-windows
	aws lambda update-function-code --function-name backend-go-lambda --zip-file fileb://function.zip

.PHONY: clean-windows
clean-windows:
	if exist bin rmdir /s /q bin
	if exist function.zip del /q function.zip

.PHONY: generate-swagger
generate-swagger:
	cd http-handler && swag init --parseDependency --parseInternal

.PHONY: lint
lint:
	go mod tidy
	go mod download
	golangci-lint run --path-prefix BackEnd --print-resources-usage --timeout 5m ./...
