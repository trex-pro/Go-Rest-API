FROM golang:1.25-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Ensure the embedded env file exists, but keep it empty to prevent baking secrets.
RUN printf "" > /src/cmd/api/.env

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN go build -trimpath -ldflags="-s -w" -o /out/project-api ./cmd/api

FROM alpine:3.21 AS runtime

RUN apk add --no-cache ca-certificates && adduser -D -H -u 10001 app

WORKDIR /app
COPY --from=build /out/project-api /app/project-api

USER app

EXPOSE 3000

ENTRYPOINT ["/app/project-api"]
