FROM alpine:3.20
RUN apk add --no-cache ca-certificates nodejs npm
# Install AI coding CLIs at runtime; no keys baked in (ADR-0005).
WORKDIR /workspace
CMD ["sleep", "infinity"]