FROM golang:1.17-buster as build-go
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o main .

FROM node:16.9.1-stretch as build-node
WORKDIR /usr/src/app
COPY . .
RUN cd assets && npm install
RUN cd assets && npm rebuild node-sass
RUN cd assets && npm run build

FROM alpine:latest
RUN apk add ca-certificates curl
WORKDIR /opt
COPY --from=build-go /go/src/app/ /bin
RUN chmod +x /bin/main
COPY --from=build-node /usr/src/app/public /opt/public
COPY --from=build-node /usr/src/app/templates /opt/templates
CMD ["/bin/main"]