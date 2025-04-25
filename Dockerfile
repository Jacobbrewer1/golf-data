FROM docker.io/ubuntu:latest

ARG COPY_LOCATION="/app"
COPY ${COPY_LOCATION} /usr/local/bin/application
ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["application"]
