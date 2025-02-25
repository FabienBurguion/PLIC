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
	go test ./http-handler -v
	docker compose down -v
