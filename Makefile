GOPATH := $(shell pwd)

all: media/flvsak


media/flvsak:
	GOPATH=$(GOPATH) go get -d $@
	GOPATH=$(GOPATH) go build $@

clean:
	GOPATH=$(GOPATH) go clean
