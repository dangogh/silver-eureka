FROM golang:1.25-alpine as builder
WORKDIR /app

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
