.PHONY: test
test:
	go test ./http-handler -v

.PHONY: build
build:
	cd http-handler && go build -o ../bin/http-handler.exe

.PHONY: docker-up
docker-up:
	docker compose up -d

.PHONY: docker-down
docker-down:
	docker compose down -d
