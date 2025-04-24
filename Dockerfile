ARG APP_NAME=app

FROM docker.io/ubuntu:latest

COPY bazel-out/k8-fastbuild/bin/cmd/${APP_NAME}/${APP_NAME}_/${APP_NAME} /usr/local/bin/application
ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["application"]
