#!/bin/bash

cd $(dirname "$0")

if [ "$UID" != "0" ]; then
  echo "Please run as root!"
  exit 1
fi

mkdir -p /etc/aspiesoft-clamav-scanner
cp -rf ./bin/* /etc/aspiesoft-clamav-scanner/
cp -f ./uninstall.sh /etc/aspiesoft-clamav-scanner/

ln -s /etc/aspiesoft-clamav-scanner/aspiesoft-clamav-download-scanner.service /etc/systemd/system/aspiesoft-clamav-download-scanner.service
ln -s /etc/aspiesoft-clamav-scanner/avscan /usr/local/bin/avscan

cd /etc/aspiesoft-clamav-scanner
go build &>/dev/null

systemctl enable aspiesoft-clamav-download-scanner.service --now
