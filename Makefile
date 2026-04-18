VERSION := $(shell cat VERSION)
LDFLAGS := -X github.com/fserg/md-to-html/internal/version.Version=$(VERSION)

.PHONY: build run test templ tailwind dev docker clean tools

build:
	go build -ldflags "$(LDFLAGS)" -o bin/md-to-html ./cmd/md-to-html

run:
	go run ./cmd/md-to-html serve

test:
	go test ./...

templ:
	templ generate

tailwind:
	@echo "tailwind target will be implemented in phase 4"

dev:
	@echo "dev target will be implemented in phase 4"

docker:
	@echo "docker target will be implemented in phase 6"

clean:
	rm -rf bin/ tmp/ web/static/dist/

tools:
	go install github.com/a-h/templ/cmd/templ@v0.3.1001
