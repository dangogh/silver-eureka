FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN apk add --no-cache build-base gcc sqlite
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# CGO_ENABLED=1 is set to enable cgo for SQLite support
RUN CGO_ENABLED=1 go build ./cmd/gather-requests

FROM alpine:3.20

RUN apk --no-cache add ca-certificates sqlite wget && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /home/appuser

COPY --from=builder /app/gather-requests /home/appuser/bin/gather-requests
RUN chown -R appuser:appuser /home/appuser

USER appuser
ENV PATH="/home/appuser/bin:${PATH}"
CMD ["gather-requests"]
