export PATH := $(GOPATH)/bin:$(PATH)
export GO111MODULE=on
LDFLAGS := -s -w

all: build

build: snps snpc

snps:
	env CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -trimpath -ldflags "$(LDFLAGS)" -o bin/snps ./cmd/snps

snpc:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/snpc ./cmd/snpc
	
clean:
	rm -f ./bin/srpc
	rm -f ./bin/srps
