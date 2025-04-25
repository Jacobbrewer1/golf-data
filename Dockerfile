ARG APP_NAME=app

FROM docker.io/ubuntu:latest

COPY "bin/${APP_NAME}" /usr/local/bin/application
ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["application"]
