.PHONE: all
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
	go test ./... -v
	docker compose down -v

.PHONY: zip-windows
zip-windows:
	powershell -ExecutionPolicy Bypass -File build.ps1

.PHONY: deploy-windows
deploy: zip-windows
	aws lambda update-function-code --function-name backend-go-lambda --zip-file fileb://function.zip
