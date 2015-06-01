#!/usr/bin/env bash
set -ue

APPNAME=${1:-"nora"}
LOGOUTPUT=${2:-"hello_world!"}
STACK=${3:-"windows2012R2"}
GREEN='\033[0;32m'
CF_COLOR='\033[0;35m'
NC='\033[0m'

echo -e "${GREEN}First, we delete the app named ${APPNAME}.${NC}"
read -p "Press [Enter] key to continue..."
echo -e "${CF_COLOR}cf d -f ${APPNAME} ${NC}"
cf d -f $APPNAME
echo -e "${CF_COLOR}cf apps ${NC}"
cf apps
echo -e "${GREEN}As we can see, there is now no app named ${APPNAME}. Next, we will push an app named ${APPNAME}.${NC}"
read -p "Press [Enter] key to continue..."
echo -e "${CF_COLOR}cf push -m 1G $APPNAME -s $STACK -b https://github.com/ryandotsmith/null-buildpack.git -p NoraPublished --no-start -f ${APPNAME} ${NC}"
cf push $APPNAME -m 1G -s $STACK -b https://github.com/ryandotsmith/null-buildpack.git -p NoraPublished --no-start
cf set-env $APPNAME DIEGO_BETA true
cf set-env $APPNAME DIEGO_RUN_BETA true
echo -e "${CF_COLOR}cf enable-diego ${APPNAME} ${NC}"
cf enable-diego $APPNAME || echo "enable diego plugin doesn't exist. follow the instructions in the README to install it"
echo -e "${CF_COLOR}cf start ${APPNAME} ${NC}"
cf start $APPNAME
echo -e "${GREEN}Now that ${APPNAME} has been pushed, we will attempt to connect to it.${NC}"
read -p "Press [Enter] key to continue..."
URL=`cf app $APPNAME | grep urls | awk '{print $2}'`
echo -e "${CF_COLOR}curl ${URL} ${NC}"
curl $URL
echo ''
echo -e "${GREEN}We can also view ${APPNAME} in the browser.${NC}"
read -p "Press [Enter] key to continue..."
open "http://${URL}"
echo -e "${GREEN}Next, we can view ${APPNAME}'s env variables. Notice the INSTANCE_INDEX variable.${NC}"
read -p "Press [Enter] key to continue..."
echo -e "${CF_COLOR}curl ${URL}/env | jsonpp ${NC}"
curl $URL/env | jsonpp
read -p "Press [Enter] key to continue..."
echo -e "${CF_COLOR}curl ${URL}/env/INSTANCE_INDEX ${NC}"
printf  "INSTANCE_INDEX:  " 
curl $URL/env/INSTANCE_INDEX
echo ""
echo -e "${GREEN}Next, we will scale ${APPNAME} to 2 instances.${NC}"
read -p "Press [Enter] key to continue..."
echo -e "${CF_COLOR}cf scale -i 2 ${APPNAME} ${NC}"
cf scale -i 2 $APPNAME
sleep 3
echo -e "${GREEN} We will now send 10 requests for INSTANCE_INDEX to ${APPNAME}. We should see responses from both instances, showing that app has been scaled.${NC}"
read -p "Press [Enter] key to continue..."
echo -e "${CF_COLOR}curl ${URL}/env/INSTANCE_INDEX ${NC}"
curl  $URL/env/INSTANCE_INDEX
echo ""
echo -e "${CF_COLOR}curl ${URL}/env/INSTANCE_INDEX ${NC}"
curl  $URL/env/INSTANCE_INDEX
echo ""
echo -e "${CF_COLOR}curl ${URL}/env/INSTANCE_INDEX ${NC}"
curl  $URL/env/INSTANCE_INDEX
echo ""
echo -e "${CF_COLOR}curl ${URL}/env/INSTANCE_INDEX ${NC}"
curl  $URL/env/INSTANCE_INDEX
echo ""
echo -e "${CF_COLOR}curl ${URL}/env/INSTANCE_INDEX ${NC}"
curl  $URL/env/INSTANCE_INDEX
echo ""
echo -e "${CF_COLOR}curl ${URL}/env/INSTANCE_INDEX ${NC}"
curl  $URL/env/INSTANCE_INDEX
echo ""
echo -e "${CF_COLOR}curl ${URL}/env/INSTANCE_INDEX ${NC}"
curl  $URL/env/INSTANCE_INDEX
echo ""
echo -e "${CF_COLOR}curl ${URL}/env/INSTANCE_INDEX ${NC}"
curl  $URL/env/INSTANCE_INDEX
echo ""
echo -e "${CF_COLOR}curl ${URL}/env/INSTANCE_INDEX ${NC}"
curl  $URL/env/INSTANCE_INDEX
echo ""
echo -e "${CF_COLOR}curl ${URL}/env/INSTANCE_INDEX ${NC}"
curl  $URL/env/INSTANCE_INDEX
echo ""
echo -e "${CF_COLOR}curl ${URL}/env/INSTANCE_INDEX ${NC}"
curl  $URL/env/INSTANCE_INDEX
echo ""
echo -e "${GREEN}Now, we will look at ${APPNAME}'s logs.${NC}"
read -p "Press [Enter] key to continue..."
echo -e "${CF_COLOR}cf logs ${APPNAME} --recent ${NC}"
cf logs $APPNAME --recent
echo -e "${GREEN}An app can write to the logs by printing to standard out. For example, we can have ${APPNAME} write to stdout using a defined endpoint.${NC}"
read -p "Press [Enter] key to continue..."
echo -e "${CF_COLOR}curl ${URL}/print/${LOGOUTPUT}${NC}"
curl  $URL/print/$LOGOUTPUT
echo
echo -e "${GREEN}Now if we grep the logs, we can see ${LOGOUTPUT} in the output."
read -p "Press [Enter] key to continue..."
echo -e "${CF_COLOR}cf logs ${APPNAME} --recent | grep ${LOGOUTPUT}${NC}"
cf logs $APPNAME --recent | grep $LOGOUTPUT
echo -e "${GREEN} Notice that the logs are categorized. Any logs outputted by ${APPNAME} with be tagged with [APP]."
read -p "Press [Enter] key to continue..."
echo -e "${GREEN} The End!"
