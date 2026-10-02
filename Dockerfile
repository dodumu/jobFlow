# ---------- Build stage ----------
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o jobflow .


# ---------- Runtime stage ----------
# ---------- Runtime stage ----------
FROM alpine:latest

WORKDIR /app

RUN addgroup -S jobflow && adduser -S jobflow -G jobflow

COPY --from=builder /app/jobflow ./jobflow
COPY --from=builder /app/templates ./templates

RUN mkdir -p /data && \
    chown -R jobflow:jobflow /data

USER jobflow

EXPOSE 8081

CMD ["./jobflow"]