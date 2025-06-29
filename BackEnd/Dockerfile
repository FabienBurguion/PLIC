FROM golang:1.24-rc

RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.56.2

WORKDIR /app

ENTRYPOINT ["golangci-lint"]
CMD ["run", "./..."]