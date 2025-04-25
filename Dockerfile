FROM docker.io/ubuntu:latest

ARG APP_NAME=api
COPY bin/${APP_NAME} /usr/local/bin/application
ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["application"]
