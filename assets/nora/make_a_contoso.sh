#!/usr/bin/env bash

set -eu

appname=${1:-"contoso"}
service=${2:-"rds"}
catalog=${3:-"ContosoUniversity2"}

if [ "x$HOST" == "x" ]; then
    echo "HOST environment variable must be set"
    exit 1
fi

if [ "x$DB_USER" == "x" ]; then
    echo "DB_USER environment variable must be set"
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
user=$DB_USER
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
cf start $appname
exit 0
