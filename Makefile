APP_NAME = consul-alerts
VERSION = latest

all: clean build

build: clean
	env CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/$(APP_NAME) .

build-local: clean
	env CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/$(APP_NAME) .

build-local-docker: clean
	docker build -t ilert/$(APP_NAME):latest .

clean:
	rm -rf ./bin

run:
	go run .
