VERSION := $(shell cat VERSION)
LDFLAGS := -X github.com/fserg/md-to-html/internal/version.Version=$(VERSION)
GOBIN := $(shell go env GOPATH)/bin
TEMPL := $(GOBIN)/templ

.PHONY: build run test templ tailwind dev docker clean tools

build:
	go build -ldflags "$(LDFLAGS)" -o bin/md-to-html ./cmd/md-to-html

run:
	go run ./cmd/md-to-html serve

test:
	go test ./...

templ:
	$(TEMPL) generate ./...

tailwind:
	mkdir -p web/static/dist
	npx tailwindcss -i web/static/src/app.css -o web/static/dist/app.css --minify

dev:
	mkdir -p web/static/dist
	sh -c 'npx tailwindcss -i web/static/src/app.css -o web/static/dist/app.css --watch & \
	TAILWIND_PID=$$!; \
	trap "kill $$TAILWIND_PID" EXIT INT TERM; \
	$(TEMPL) generate --watch --proxy=http://localhost:8080 --cmd="go run ./cmd/md-to-html serve"'

docker:
	@echo "docker target will be implemented in phase 6"

clean:
	rm -rf bin/ tmp/ web/static/dist/

tools:
	go install github.com/a-h/templ/cmd/templ@v0.3.1001
	go install github.com/templui/templui/cmd/templui@latest
