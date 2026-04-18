# syntax=docker/dockerfile:1.7

FROM debian:bookworm-slim AS tailwind

RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates curl \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /src

ARG TARGETARCH
ARG TAILWIND_VERSION=v3.4.17

RUN case "$TARGETARCH" in \
	amd64) tailwind_arch='x64' ;; \
	arm64) tailwind_arch='arm64' ;; \
	*) echo "unsupported TARGETARCH: $TARGETARCH" >&2; exit 1 ;; \
	esac \
	&& curl -fsSL -o /usr/local/bin/tailwindcss \
		"https://github.com/tailwindlabs/tailwindcss/releases/download/${TAILWIND_VERSION}/tailwindcss-linux-${tailwind_arch}" \
	&& chmod +x /usr/local/bin/tailwindcss

COPY tailwind.config.js ./
COPY web/ ./web/
COPY internal/ui/ ./internal/ui/

RUN mkdir -p web/static/dist \
	&& tailwindcss \
		-c tailwind.config.js \
		-i web/static/src/app.css \
		-o web/static/dist/app.css \
		--minify

FROM golang:1.24-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates git
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1001

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=tailwind /src/web/static/dist/app.css ./web/static/dist/app.css

RUN templ generate ./... \
	&& CGO_ENABLED=0 GOOS=linux go build \
		-trimpath \
		-ldflags="-s -w -X github.com/fserg/md-to-html/internal/version.Version=$(cat VERSION)" \
		-o /out/md-to-html \
		./cmd/md-to-html

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/md-to-html /md-to-html

EXPOSE 8080

USER nonroot

ENTRYPOINT ["/md-to-html", "serve"]
