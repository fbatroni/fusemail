#!/bin/sh

VENDOR='fusemail'
PROJECT_NAME='em-service-usage-importer'
PROD_ENV_PATH="/etc/${VENDOR}/${PROJECT_NAME}"

chown nobody:nogroup $PROD_ENV_PATH/${PROJECT_NAME}-prod.env
chmod 0600 $PROD_ENV_PATH/${PROJECT_NAME}-prod.env
