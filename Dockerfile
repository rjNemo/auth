# syntax=docker/dockerfile:1.7
ARG GO_VERSION=1.25.1

FROM golang:${GO_VERSION}-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

RUN apk add --no-cache build-base git

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -trimpath -ldflags="-s -w" -o /out/auth-server ./cmd/server

FROM gcr.io/distroless/base-nonroot:latest

WORKDIR /app

COPY --from=build /out/auth-server ./auth-server

USER nonroot:nonroot
EXPOSE 8000

ENTRYPOINT ["/app/auth-server"]
