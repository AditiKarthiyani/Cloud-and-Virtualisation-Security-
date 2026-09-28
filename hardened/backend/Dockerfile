FROM golang:1.25-bookworm AS builder

RUN apt-get update && apt-get upgrade -y &&  apt-get install -y \
    git \
    build-essential && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /server ./

FROM postgres:18.4
ARG POSTGRES_DB=exercise
ENV POSTGRES_DB=${POSTGRES_DB}

# non-root user for GO API server
RUN groupadd -r appgroup && useradd -r -g appgroup appuser

COPY --from=builder /server /server
COPY start.sh /start.sh
RUN chmod +x /start.sh /server && \
    chown appuser:appgroup /server /start.sh

EXPOSE 8081 
USER appuser
HEALTHCHECK --interval=50s --timeout=5s --retries=5 \
  CMD pg_isready -h "$DB_HOST" -p "$DB_PORT" || exit 1
ENTRYPOINT ["/start.sh"]
