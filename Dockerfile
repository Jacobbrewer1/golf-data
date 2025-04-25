FROM docker.io/ubuntu:latest

COPY bin/api /usr/local/bin/application
ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["application"]
