.PHONY: run
run:
	go build -ldflags '-w -s' -a -o ./bin/app ./cmd/logistic_optimizer/main.go && HTTP_ADDR=8080 ./bin/app

.PHONY: lint
lint:
	golangci-lint run