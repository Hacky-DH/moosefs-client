version?=0.0.1
pkg=mfscli
src=main.go
flag+=-X main.versionstr=v$(version)$(shell date +.%Y%m%d.%H%M%S)
tarname=$(pkg)-$(shell uname -s)-$(shell uname -p).tar.gz

all: build

deps:
        go get github.com/golang/glog
        go get github.com/google/subcommands
        go get github.com/Hacky-DH/moosefs-client

fmt:
        go fmt $(src)

build: deps fmt
        go build -o $(pkg) -ldflags "$(flag)" $(src)
        strip -s $(pkg)

install: deps fmt
        go install -o $(pkg) -ldflags "$(flag)" $(src)

tar: build
        tar -czf $(tarname) $(pkg)

test: deps fmt
        go test -v

clean:
        /bin/rm -rf $(pkg) $(tarname)

.PHONY: all deps fmt build install tar test clean
