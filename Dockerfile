ARG APP_NAME=app

FROM docker.io/golang:alpine as build

ARG APP_NAME

WORKDIR /build

COPY . /build/
RUN go build -o application ./cmd/${APP_NAME}

FROM docker.io/ubuntu:latest

COPY --from=build /build/application /usr/local/bin/application
ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["application"]
