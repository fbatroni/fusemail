ARG registry=docker-registry.electric.net:10080
ARG build_env_version=latest
ARG debian_version=latest

# checkout base image content in fm-lib-common-dockerfile repo
FROM ${registry}/fusemail/fm-utility-go-build:${build_env_version} AS builder
ARG project_name=project
ENV PROJECT=$project_name
WORKDIR /go/src/bitbucket.org/fusemail/${PROJECT}/
COPY . .
ARG GOOS=linux
RUN GOOS="${GOOS}" make build

FROM ${registry}/debian/debian:${debian_version}
RUN apt-get update \
    && apt-get -y upgrade \
    && apt-get install -y ca-certificates
ARG project_name=project
ENV PROJECT=$project_name
WORKDIR /etc/fusemail/${PROJECT}/
COPY --from=builder /go/src/bitbucket.org/fusemail/$PROJECT/conf/$PROJECT-*.env ./
WORKDIR /usr/local/fusemail/${PROJECT}
COPY --from=builder /go/src/bitbucket.org/fusemail/$PROJECT/build/$PROJECT .
RUN echo "/usr/local/fusemail/$PROJECT/$PROJECT \$@" > ./entrypoint.sh
ENTRYPOINT ["/bin/sh", "./entrypoint.sh"]