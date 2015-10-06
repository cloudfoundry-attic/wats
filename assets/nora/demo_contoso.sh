#!/usr/bin/env bash

set -eu

appname=${1:-"contoso"}
service=${2:-"rds"}
catalog=${3:-"ContosoUniversity2"}
GREEN='\033[0;32m'
CF_COLOR='\033[0;35m'
NC='\033[0m'

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

echo -e "${GREEN}First, we delete the app named ${appname}.${NC}"
read -p "Press [Enter] key to continue..."
echo -e "${CF_COLOR}cf d -f ${appname} ${NC}"
cf d -f $appname
echo -e "${CF_COLOR}cf delete-service -f ${service} ${NC}"
cf delete-service -f $service
echo -e "${CF_COLOR}cf delete-security-group -f ${service} ${NC}"
cf delete-security-group -f $service

echo -e "${CF_COLOR}cf apps ${NC}"
cf apps
echo -e "${GREEN}As we can see, there is now no app named ${appname}. Next, we will push an app named ${appname}.${NC}"
echo -e "${GREEN}Note that this is a completely stock version of the Contoso University project (found at https://code.msdn.microsoft.com/ASPNET-MVC-Application-b01a9fe8).${NC}"
read -p "Press [Enter] key to continue..."
buildpack=https://github.com/ryandotsmith/null-buildpack.git

echo -e "${CF_COLOR}cf push ${appname} -s windows2012R2 -b ${buildpack} -p contosopublished --no-start ${NC}"
cf push $appname -s windows2012R2 -b $buildpack -p contosopublished --no-start
echo -e "${CF_COLOR}cf enable-diego ${appname} ${NC}"
cf enable-diego $appname

host=$HOST
user=$DB_USER
pass=$PASS

echo -e "${GREEN}Now that ${appname} has been pushed, we will create a user provided service..${NC}"
echo -e "${GREEN}A user provided service with a name, connectionString, and providerName will automatically be spliced into the app's web.config.${NC}"
read -p "Press [Enter] key to continue..."
credentials=$(cat <<EOF
{
  "name":"SchoolContext",
  "connectionString":"Data Source=$host;Initial Catalog=$catalog;User ID=$user;Password=$pass",
  "providerName":"System.Data.SqlClient"
}
EOF
       )

echo -e "${CF_COLOR}cf cups ${service} -p ${credentials} ${NC}"
cf cups $service -p "$credentials"

echo -e "${CF_COLOR}cf bind-service ${appname} ${service} ${NC}"
cf bind-service $appname $service

echo -e "${GREEN}We create a security group to ensure that the app can communicate outbound with RDS.${NC}"
read -p "Press [Enter] key to continue..."
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
echo -e "${CF_COLOR} ${service}-security-group.json: "
cat $service-security-group.json
echo -e "${NC}"

org=`cf target | grep Org | awk '{print $2}'`
space=`cf target | grep Space | awk '{print $2}'`

echo -e "${CF_COLOR}cf create-security-group ${service} ${service}-security-group.json ${NC}"
cf create-security-group $service $service-security-group.json
rm $service-security-group.json
echo -e "${CF_COLOR}cf bind-security-group ${service} ${org} ${space} ${NC}"
cf bind-security-group $service $org $space
echo -e "${GREEN}Now, we will start ${appname}. ${NC}"
read -p "Press [Enter] key to continue..."
echo -e "${CF_COLOR}cf start ${appname} ${NC}"
cf start $appname
URL=`cf app ${appname} | grep urls | awk '{print $2}'`
echo -e "${GREEN}[tour of contoso]${NC}"
open "http://${URL}"
read -p "Press [Enter] key to continue..."
echo -e "${GREEN}Next, we will scale ${appname} to 2 instances.${NC}"
read -p "Press [Enter] key to continue..."
echo -e "${CF_COLOR}cf scale -i 2 ${appname} ${NC}"
cf scale -i 2 $appname
sleep 3
echo -e "${CF_COLOR}cf apps ${NC}"
cf apps
echo -e "${GREEN}If we added an implementor of Contoso's ILogger that printed to stdout and stderr instead, we would be able to see its logs using 'cf logs ${appname}'.${NC}"
read -p "Press [Enter] key to continue..."
echo -e "${GREEN}The End!${NC}"
exit 0
