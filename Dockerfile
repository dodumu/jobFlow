# ---------- Build stage ----------
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o jobflow .


# ---------- Runtime stage ----------
FROM alpine:latest

WORKDIR /app

RUN addgroup -S jobflow && adduser -S jobflow -G jobflow

COPY --from=builder /app/jobflow ./jobflow
COPY --from=builder /app/templates ./templates

# /data holds everything that must survive restarts: the database and uploads.
RUN mkdir -p /data/uploads && \
    chown -R jobflow:jobflow /data

ENV PORT=8081 \
    DATABASE_PATH=/data/database.db \
    UPLOAD_PATH=/data/uploads

VOLUME ["/data"]

USER jobflow

EXPOSE 8081

CMD ["./jobflow"]