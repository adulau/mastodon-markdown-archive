FROM golang:1.23.0-alpine3.20 AS build

WORKDIR /app
COPY go.mod go.sum .
RUN go mod download

COPY . .
ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go build -o mastodon-markdown-archive

FROM alpine:3.20
RUN apk add --no-cache ca-certificates

LABEL org.opencontainers.image.title="Mastodon Markdown Archive"
LABEL org.opencontainers.image.description="Archive Mastodon posts as markdown files"
LABEL org.opencontainers.image.vendor="Gabriel Garrido"
LABEL org.opencontainers.image.licenses=MIT
LABEL org.opencontainers.image.url=https://git.garrido.io/gabriel/mastodon-markdown-archive
LABEL org.opencontainers.image.source=https://git.garrido.io/gabriel/mastodon-markdown-archive
LABEL org.opencontainers.image.documentation=https://git.garrido.io/gabriel/mastodon-markdown-archive

COPY --from=build /app/mastodon-markdown-archive /usr/bin/mastodon-markdown-archive
ENTRYPOINT ["/usr/bin/mastodon-markdown-archive"]
