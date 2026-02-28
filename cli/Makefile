VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

BINARY  := sharepwd
LDFLAGS := -s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(DATE)

PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

.PHONY: build clean all test checksums

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY) .

all: clean
	@mkdir -p bin
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		output="bin/$(BINARY)-$$os-$$arch$$ext"; \
		echo "Building $$output..."; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -ldflags="$(LDFLAGS)" -o $$output . || exit 1; \
	done

test:
	go test -v -race ./...

checksums:
	@cd bin && sha256sum $(BINARY)-* > checksums.txt
	@echo "Checksums written to bin/checksums.txt"

clean:
	rm -rf bin/
