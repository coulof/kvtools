# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /src
RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/kvtools main.go

# Runtime stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /bin/kvtools /usr/local/bin/kvtools
COPY --from=builder /bin/kvtools /usr/local/bin/kubectl-kvtools

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/kvtools"]
