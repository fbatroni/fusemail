#!/bin/sh

VENDOR='fusemail'
PROJECT_NAME='em-service-usage-importer'
ENVDIR="/etc/${VENDOR}/${PROJECT_NAME}/env"
INSTALL_DIR="/usr/local/${VENDOR}/${PROJECT_NAME}"

if [ ! -d $ENVDIR ]; then
    mkdir -p $ENVDIR
fi

if [ ! -d $INSTALL_DIR ]; then
    mkdir -p $INSTALL_DIR
fi
