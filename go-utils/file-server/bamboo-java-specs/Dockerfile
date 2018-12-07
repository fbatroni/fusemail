ARG registry=docker-registry.electric.net:10080
ARG build_env_version=latest

FROM ${registry}/fusemail/fm-utility-java-build:${build_env_version}
WORKDIR /java/src/bamboo-specs/
COPY . .

ENTRYPOINT [ "mvn" ]