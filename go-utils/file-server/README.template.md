# README

This service (fm-app-go-template) is ...

Further details in [Architecture](https://fusemail.atlassian.net/wiki/display/DEMO).

Based from [Go App Template](https://bitbucket.org/fusemail/fm-app-go-template).

## Available Endpoints

### System

* [Health](/health)
* [Metrics](/metrics)
* [System Variables](/sys)
* [Version](/version)

### Documentation

* [API Documentation](/api)

### Logging and Debugging

* [Text Log](/log)
* [JSON Log](/log?format=json)
* [Debugging](/debug/pprof/)

Generic details in [Commons Application endpoints](https://bitbucket.org/fusemail/fm-lib-commons-golang#markdown-header-application-endpoints).

## Options

Invoke help for options

    # make build && ./build/fm-app-go-template --help

Generic details in [Commons Configuration](https://bitbucket.org/fusemail/fm-lib-commons-golang#markdown-header-configuration).

If you don't have Golang installed in your environment, then you can also use [Docker](https://docs.docker.com/install/):

    # make docker-run-help

## Deployment

### With Debian Package

Build for production

    # make

_Side note_: if you experience problems, refer to [Commons Setup](https://bitbucket.org/fusemail/fm-lib-commons-golang#markdown-header-setup).

This creates a _/build_ directory with the following files (binaries and debian package)

* fm-app-go-template
* fm-app-go-template-VERSION
* fm-app-go-template_VERSION_amd64.deb

Optionally copy ```fm-app-go-template-dev.env``` environment file to _/build_ directory, rename _dev_ to the desired environment key, and edit as necessary.

Other _make_ options

    # make help

Run after _make_

    # cd build
    ~/build$ ./fm-app-go-template

Optionally suppply ```-e ENV``` and ```--environment-path EnvPath``` flags to load environment variables from the following paths

* build/fm-app-go-template-{ENV}.env
* {EnvPath}/fm-app-go-template-{ENV}.env

Generic details in [Commons Deployment](https://bitbucket.org/fusemail/fm-lib-commons-golang#markdown-header-deployment).

### With Docker

Build runnable Docker image:

    # make docker

Push Docker image to registry:

    # make docker-push

The built binary and default environment files are stored within the Docker image at the following location:

    /usr/local/fusemail/fm-app-go-template/
        fm-app-go-template
    /etc/fusemail/fm-app-go-template/
        fm-app-go-template-dev.env
        fm-app-go-template-prod.env

### With Nomad

Nomad deployment requires Docker container, so the above step is a prerequisite.

The following command builds jobspec for each datacenter and uploads to artifact server, but can only be run by Bamboo plans:

    # make deploy

## Build for development

Build and run for development

    # make build && ./build/fm-app-go-template

Generic details in [Commons Run](https://bitbucket.org/fusemail/fm-lib-commons-golang#markdown-header-run).

If you don't want to install Golang in your own environment, then you can also use [Docker](https://docs.docker.com/install/):

    # make docker-run       // make sure you change the default port number in Makefile if it's not 8080
    # make docker-run-stop  // stops the test run container

## Test

Convey tests

    # go test -v

Generic details in [Commons Test](https://bitbucket.org/fusemail/fm-lib-commons-golang#markdown-header-test).

If you don't want to install Golang in your own environment, then you can also use [Docker](https://docs.docker.com/install/):

    # make docker-test