# ===============================================
# build stage
# ===============================================
FROM golang:alpine AS build-env
RUN apk add --update \
    git \
  && rm -rf /var/cache/apk/*
RUN go get github.com/RedisLabs/sentinel_tunnel
ADD . /src
RUN cd /src && go build -o sentinel_tunnel

# ===============================================
# final stage
# ===============================================
FROM alpine
WORKDIR /app
RUN mkdir /app/logs && touch /app/logs/output.log
VOLUME /app/logs
COPY --from=build-env /src/sentinel_tunnel /app/
CMD ["/app/sentinel_tunnel",  "config.json",  "logs/output.log"]
