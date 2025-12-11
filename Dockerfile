FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN apk add --no-cache build-base gcc sqlite
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# CGO_ENABLED=1 is set to enable cgo for SQLite support
RUN CGO_ENABLED=1 go build ./cmd/gather-requests

FROM alpine:3.20

RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/

COPY --from=builder /app/gather-requests /root/bin/
CMD ["gather-requests"]
