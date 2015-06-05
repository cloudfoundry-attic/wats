#!/usr/bin/env bash

set -eu

# secrets-hurt.c9a9owfndn3b.us-east-1.rds.amazonaws.com
# secrets
# 912ff38a-d244-487b-b85a-db9274f0c7f2

function disable_ssh {
    space=`cf target | grep Space | awk '{print $2}' | tr -d '\n'`

    cf curl /v2/apps/$(cf app $1 --guid) -X PUT -d '{"enable_ssh": false}'
    # cf curl /v2/spaces/$(cf space $space --guid) -X PUT -d '{"allow_ssh": false}'
}

appname=${1:-"contoso"}
service=${2:-"rds"}
catalog=${3:-"ContosoUniversity2"}

if [ "x$HOST" == "x" ]; then
    echo "HOST environment variable must be set"
    exit 1
fi

if [ "x$USER" == "x" ]; then
    echo "USER environment variable must be set"
    exit 1
fi

if [ "x$PASS" == "x" ]; then
    echo "PASS environment variable must be set"
    exit 1
fi

cf d -f $appname
cf delete-service -f $service
cf delete-security-group -f $service

buildpack=https://github.com/ryandotsmith/null-buildpack.git

cf push $appname -m 2g -s windows2012R2 -b $buildpack -p contosopublished --no-start
cf enable-diego $appname

host=$HOST
user=$USER
pass=$PASS

credentials=$(cat <<EOF
{
  "name":"SchoolContext",
  "connectionString":"Data Source=$host;Initial Catalog=$catalog;User ID=$user;Password=$pass",
  "providerName":"System.Data.SqlClient"
}
EOF
       )

cf cups $service -p "$credentials"

cf bind-service $appname $service

ip=`nslookup $host | grep Address | tail -n1 | awk '{print $2}'`
cat <<EOF > $service-security-group.json
[
  {
    "protocol": "tcp",
    "destination": "$ip",
    "ports": "1433"
  }
]
EOF

org=`cf target | grep Org | awk '{print $2}'`
space=`cf target | grep Space | awk '{print $2}'`

cf create-security-group $service $service-security-group.json
rm $service-security-group.json
cf bind-security-group $service $org $space
disable_ssh $appname
cf start $appname
exit 0
