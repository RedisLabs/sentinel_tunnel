FROM scratch

COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY sentinel_tunnel /

ENTRYPOINT ["/sentinel_tunnel", "/config.json"]
