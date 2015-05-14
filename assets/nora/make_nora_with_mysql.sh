#!/usr/bin/env bash

set -e

./make_a_nora

if [ "x${NORA_PASS}" == "x" ]; then
    echo "You must provide NORA_PASS env variables"
    exit 1
fi

if [ "x${ROOT_PASS}" != "x" ]; then
    cat <<EOF | mysql -uroot -p${ROOT_PASS}
CREATE USER 'nora'@'%' identified by 'nora';
GRANT ALL ON *.* TO 'nora'@'%';
CREATE USER 'nora'@'localhost' identified by 'nora';
GRANT ALL ON *.* TO 'nora'@'localhost';
EOF
else
    echo "ROOT_PASS wasn't set, won't create nora user in mysql"
fi

cf cups mysql -p "{\"username\":\"nora\",\"password\":\"${NORA_PASS}\",\"host\":\"172.16.234.1\"}"
cf bind-service nora mysql

cat <<EOF > mysql-security-group.json
[
  {
    "protocol": "tcp",
    "destination": "172.16.234.1",
    "ports": "3306"
  }
]
EOF
cf create-security-group mysql mysql-security-group.json
rm mysql-security-group.json
cf bind-security-group mysql org space
cf restage nora
exit 0
