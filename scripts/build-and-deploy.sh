#!/bin/sh

set -e

if [[ $GRAFANA_PLUGIN_PATH = "" ]] || [[ $PG_MONITORING_PATH = "" ]]; then
    echo "export GRAFANA_PLUGIN_PATH && PG_MONITORING_PATH && PLUGIN_NAME"
    echo "e.g. export GRAFANA_PLUGIN_PATH /path/to/grafana/plugins"
    exit 1
fi;

PLUGIN_NAME=aws-athena-datasource-plugin

npm run dev 
GOOS=linux GOARCH=amd64 go build -o ./dist/aws-athena-datasource-plugin_linux_amd64 ./backend
GOOS=darwin GOARCH=amd64 go build -o ./dist/aws-athena-datasource-plugin_darwin_amd64 ./backend
rm -rf $GRAFANA_PLUGIN_PATH/${PLUGIN_NAME}/
mkdir ${GRAFANA_PLUGIN_PATH}/${PLUGIN_NAME}/
cp -rf dist ${GRAFANA_PLUGIN_PATH}/${PLUGIN_NAME}/
cd $PG_MONITORING_PATH
docker-compose down
docker-compose build
docker-compose up -d