# Build
FROM golang:1.23-alpine AS build

WORKDIR /app

# COPY go.mod go.sum ./ # no external docdependecy at the moment
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/aistream-proxy .

# Runtime
FROM alpine:latest

WORKDIR /app

COPY --from=build /app/bin/aistream-proxy .

RUN apk --no-cache add ca-certificates tzdata

ENTRYPOINT ["/app/aistream-proxy"]
