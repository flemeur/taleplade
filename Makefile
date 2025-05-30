.PHONY: default build docker static clean frontend

default: build

build:
	go build -trimpath -o ./ ./cmd/...

docker:
	docker build -t taleplade .

static:
	CGO_ENABLED=0 go build -a -trimpath -tags netgo -ldflags '-s -w -extldflags "-static"' -o ./ ./cmd/...

clean:
	rm -f ./taleplade
