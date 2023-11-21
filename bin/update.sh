#!/bin/bash

cd $(dirname "$0")
dir="$PWD"

mkdir ./tmp
cd ./tmp

wget https://github.com/AspieSoft/linux-clamav-download-scanner/archive/master.zip
unzip master.zip

cp -rf linux-clamav-download-scanner-master/bin/* /etc/aspiesoft-clamav-scanner/

cd "$dir"
rm -rf ./tmp

go build &>/dev/null

systemctl daemon-reload
systemctl restart aspiesoft-clamav-download-scanner.service
