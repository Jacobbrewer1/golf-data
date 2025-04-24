ARG BINARY_LOCATION=""

FROM docker.io/ubuntu:latest

COPY ${BINARY_LOCATION} /usr/local/bin/application
ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["application"]
