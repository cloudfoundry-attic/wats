#!/usr/bin/env bash
set -ue

APPNAME=${1:-"nora"}
STACK=${2:-"windows2012R2"}
GREEN='\033[0;32m'
NC='\033[0m'

cf d -f $APPNAME
echo -e "${GREEN}showing ${APPNAME} is not deployed...${NC}"
cf apps
read -p "Press [Enter] key to continue..."
echo -e "${GREEN}pushing an app...${NC}"
cf push $APPNAME -s $STACK -b java_buildpack -p NoraPublished --no-start
cf set-env $APPNAME DIEGO_BETA true
cf set-env $APPNAME DIEGO_RUN_BETA true
cf enable-diego $APPNAME || echo "enable diego plugin doesn't exist. follow the instructions in the README to install it"
cf start $APPNAME
echo -e "${GREEN}app pushed${NC}"
read -p "Press [Enter] key to continue..."
echo -e "${GREEN}showing app is running...${NC}"
URL=`cf app $APPNAME | grep urls | awk '{print $2}'`
curl $URL
echo ''
read -p "Press [Enter] key to continue..."
echo -e "${GREEN}outputing env variables...${NC}"
curl $URL/env | jsonpp
read -p "Press [Enter] key to continue..."
echo -e "${GREEN}scaling app to 2 instances...${NC}"
cf scale -i 2 $APPNAME
sleep 3
read -p "Press [Enter] key to continue..."
echo -e "${GREEN}printing app instances given 10 requests...${NC}"
curl  $URL/env/INSTANCE_INDEX
echo ""
curl  $URL/env/INSTANCE_INDEX
echo ""
curl  $URL/env/INSTANCE_INDEX
echo ""
curl  $URL/env/INSTANCE_INDEX
echo ""
curl  $URL/env/INSTANCE_INDEX
echo ""
curl  $URL/env/INSTANCE_INDEX
echo ""
curl  $URL/env/INSTANCE_INDEX
echo ""
curl  $URL/env/INSTANCE_INDEX
echo ""
curl  $URL/env/INSTANCE_INDEX
echo ""
curl  $URL/env/INSTANCE_INDEX
echo ""
curl  $URL/env/INSTANCE_INDEX
echo ""
read -p "Press [Enter] key to continue..."
echo -e "${GREEN}printing logs of ${APPNAME}...${NC}"
cf logs $APPNAME --recent
