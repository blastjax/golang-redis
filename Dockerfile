# Build a static golang-redis binary, then copy it into a small image
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY internal ./internal
RUN CGO_ENABLED=0 go build -o /golang-redis .

FROM alpine:3.22
COPY --from=build /golang-redis /usr/local/bin/golang-redis
USER nobody
CMD ["golang-redis", "serve"]
