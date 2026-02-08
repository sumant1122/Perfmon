.PHONY: build run test clean

build:
	go build -ldflags "-X main.version=dev" -o perfmon .

run:
	go run .

test:
	go test ./...

clean:
	rm -f perfmon
