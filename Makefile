.PHONY: build run test clean

build:
	go build -o perfmon .

run:
	go run .

test:
	go test ./...

clean:
	rm -f perfmon
