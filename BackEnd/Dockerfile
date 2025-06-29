FROM golang:1.23

RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.56.2

WORKDIR /app

COPY . .

RUN go mod tidy && go mod download

ENTRYPOINT ["golangci-lint"]
CMD ["run", "./..."]