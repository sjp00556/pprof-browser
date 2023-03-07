.PHONY: all build server clean test mock

default: all

all: build

build:
	rm -rf build
	mkdir build
	go build -o ./build/pprof-browser ./cmd/*.go
	cp -rpf static ./build/
