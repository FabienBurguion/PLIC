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
	cd http-handler && powershell -Command "& { $env:GOOS='linux'; $env:GOARCH='amd64'; go build -o ../bootstrap }" 2> nul && cd .. && powershell -Command "Compress-Archive -Path bootstrap -DestinationPath function.zip -Force"
