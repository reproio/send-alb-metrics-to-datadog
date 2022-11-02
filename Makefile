
.PHONY: build
build:
	go get && go build -o main .

.PHONY: build-image
build-image:
	docker build -t ghcr.io/reproio/send-alb-metrics-to-datadog .

.PHONY: push-image
push-image:
	docker push ghcr.io/reproio/send-alb-metrics-to-datadog:latest

.PHONY: test
test:
	go test -race -v ./...
