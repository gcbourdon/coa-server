# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /build

# Copy module files and download deps first (cache layer)
COPY coa-server/go.mod coa-server/go.sum ./
RUN go mod download

# Copy server source and build
COPY coa-server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /coa-server .

# Runtime stage
FROM alpine:3.21
WORKDIR /app

COPY --from=builder /coa-server ./coa-server
COPY cards/ ./cards/

ENV CARDS_DIR=/app/cards
EXPOSE 8080

ENTRYPOINT ["./coa-server"]
