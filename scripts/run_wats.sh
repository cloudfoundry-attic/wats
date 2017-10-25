#!/usr/bin/env bash

set -ex

if [ -f "$1" ]; then
  CONFIG_FILE=$PWD/$1
else
  CONFIG_FILE=`mktemp -t watsXXXXX`
  trap "rm -f $CONFIG_FILE" EXIT

: ${API:?"Must set api url (e.g. api.10.244.0.34.xip.io)"}
: ${ADMIN_USER:?"Must set admin username (e.g. admin)"}
: ${ADMIN_PASSWORD:?"Must set admin password (e.g. admin)"}
: ${APPS_DOMAIN:?"Must set app domain url (e.g. 10.244.0.34.xip.io)"}
: ${SOCKET_ADDRESS_FOR_SECURITY_GROUP_TEST:?"Must set address [ip address of Diego ETCD cluster or BOSH Director] (e.g. 10.244.16.2:4001 or 10.0.0.6:25555)"}
: ${NUM_WIN_CELLS:?"Must provide the number of windows cells in this deploy (e.g. 2)"}
: ${CONSUL_MUTUAL_TLS:=false}
: ${HTTP_HEALTHCHECK:=false}
: ${TEST_TASK:=false}
: ${SKIP_SSH:=false}
: ${STACK:=windows2012R2}
: ${CREDHUB_MODE:=none}

cat > $CONFIG_FILE <<HERE
{
  "api": "$API",
  "admin_user": "$ADMIN_USER",
  "admin_password": "$ADMIN_PASSWORD",
  "apps_domain": "$APPS_DOMAIN",
  "secure_address": "$SOCKET_ADDRESS_FOR_SECURITY_GROUP_TEST",
  "num_windows_cells": $NUM_WIN_CELLS,
  "skip_ssl_validation": true,
  "consul_mutual_tls": $CONSUL_MUTUAL_TLS,
  "http_healthcheck": $HTTP_HEALTHCHECK,
  "test_task": $TEST_TASK,
  "skip_ssh": $SKIP_SSH,
  "stack": "$STACK",
  "credhub_mode": "$CREDHUB_MODE"
}
HERE
fi

pushd `dirname $0`
if [[ "$(uname)" = "Darwin" ]]; then
  ln -sf cf-darwin ../bin/cf
else
  ln -sf cf-linux ../bin/cf
fi
popd

export PATH=$PWD/../bin:$PATH

uname_s=$(uname -s | cut -d_ -f1)
win_uname="MINGW32"
if [ "x${uname_s}" == "x${win_uname}" ]; then
    ginkgo_args="-noColor"
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"

export GOPATH=$DIR
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH
export GO15VENDOREXPERIMENT=1
export CF_DIAL_TIMEOUT=30

go install wats/vendor/github.com/onsi/ginkgo/ginkgo

shift || true
CONFIG=$CONFIG_FILE ginkgo ${ginkgo_args} -r -slowSpecThreshold=120 $@ $DIR
