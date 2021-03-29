FROM golang:1.16.1-alpine as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd/bot
RUN apk update && apk add g++
RUN go build -o /app/ooops

FROM alpine:latest
LABEL maintainer="Bogdan Kosarevskyi <bogdan.kosarevskyi@gmail.com>"
LABEL vendor="1node"
LABEL description="Ooops-bot"
COPY --from=build /app/ooops /app/ooops
COPY --from=build /app/data /app/data
WORKDIR /tmp
RUN apk update && \
    wget https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz && \
    tar xzvf ./migrate.linux-amd64.tar.gz && \
    mv migrate.linux-amd64 /usr/local/bin/migrate && \
    rm migrate.linux-amd64.tar.gz
WORKDIR /app
VOLUME /app/data
CMD ["/app/ooops"]