.PHONY: test bin docker debug
default:
	$(MAKE) all
test:
	CGO_ENABLED=1 go test -timeout 30m -race -short ./...
test-all:
	CGO_ENABLED=1 go test -timeout 30m -race ./...
check:
	$(MAKE) test
bin:
	./scripts/build.sh
bin-race:
	./scripts/build.sh -race
docker:
	./scripts/build_docker.sh
qa: bin qa-common

#debug versions for remote debugging with delve
bin-debug:
	./scripts/build.sh -debug
docker-debug:
	./scripts/build_docker.sh -debug
qa-debug: bin-debug qa-common

qa-common:
	# regular qa steps (can run directly on code)
	scripts/qa/gofmt.sh
	scripts/qa/go-generate.sh
	scripts/qa/ineffassign.sh
	scripts/qa/misspell.sh
	scripts/qa/gitignore.sh
	scripts/qa/unused.sh
	scripts/qa/vendor.sh
	scripts/qa/vet-high-confidence.sh
	# qa-post-build steps minus stack tests
	scripts/qa/docs.sh

all:
	$(MAKE) bin
	$(MAKE) docker
	$(MAKE) qa

debug:
	$(MAKE) bin-debug
	$(MAKE) docker-debug
	$(MAKE) qa-debug

clean:
	rm build/*
	rm scripts/build/*
