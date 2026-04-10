
.PHONY: all build install run test bundle-create bundle-clone help

BINARY := gitfame
PKG := ./gitfame/cmd/gitfame

all: build

build:
	cd $(PKG) && go build -o $(BINARY) .

install:
	go install $(PKG)/...

run: build
	$(PKG)/$(BINARY) --repository=. --extensions='.go,.md' --order-by=lines

test:
	go test -v ./gitfame/test/integration/... -count=1

help:
	@echo "Makefile targets:"
	@echo "  make build         - build gitfame binary"
	@echo "  make install       - install to $$GOPATH/bin"
	@echo "  make run           - build and run example (uses --repository=. --extensions='.go,.md')"
	@echo "  make test          - run integration tests"
	@echo "  make bundle-create - create git bundle (BUNDLE=my.bundle)"
	@echo "  make bundle-clone  - clone bundle (BUNDLE_PATH=/path/to/bundle)"
	@echo "  make help"
