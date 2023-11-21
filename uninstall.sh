#!/bin/bash

cd $(dirname "$0")

if [ "$UID" != "0" ]; then
  echo "Please run as root!"
  exit 1
fi

systemctl disable aspiesoft-clamav-download-scanner.service --now

rm -f /etc/systemd/system/aspiesoft-clamav-download-scanner.service
rm -f /usr/local/bin/avscan

rm -rf /etc/aspiesoft-clamav-scanner
