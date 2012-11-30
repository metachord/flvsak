GOPATH := $(shell pwd)

all: media/flvsak


media/flvsak:
	GOPATH=$(GOPATH) go get $@
	GOPATH=$(GOPATH) go build $@

