.PHONY: all build test shorttest update-readme clean

all: build

build:
	go build

test:
	go test -v

shorttest:
	go test -v -short

snapshot:
	goreleaser --snapshot --skip-publish --rm-dist

update-readme:
	tools/update-readme.sh
	doctoc README.md

clean:
	@- $(RM) ets
