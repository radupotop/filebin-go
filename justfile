[default]
all: build test

build:
    go build -v -o filebin ./

test:
    go test -v ./tests
