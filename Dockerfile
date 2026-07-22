# Stage 1: Build
FROM golang:1.25-alpine AS builder

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /adms ./cmd/adms

# Stage 2: Minimal runtime
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata wget

ENV TZ=Asia/Jakarta

COPY --from=builder /adms /adms

# Copy templates and public files
COPY templates/ /templates/
COPY public/ /public/
COPY config/config.yaml /config/config.yaml

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O- http://localhost:8081/health || exit 1

ENTRYPOINT ["/adms"]
CMD ["-config", "/config/config.yaml"]
