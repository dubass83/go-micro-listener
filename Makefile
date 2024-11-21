.PHONY: *

docker_up:
	limactl start docker

docker_down: 
	limactl stop docker

docker_build:
	docker build -t go-micro-mailer -f Dockerfile .

docker_build_simple: build
	docker build -t go-micro-mailer -f Dockerfile.simple .

test:
	go test -v -cover -count=1 -short ./...

server:
	go run main.go

build:
	go build -o main main.go