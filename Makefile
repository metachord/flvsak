GOPATH := $(shell pwd)

all: media/flvtag


media/flvtag:
	GOPATH=$(GOPATH) go get $@
	GOPATH=$(GOPATH) go build $@

