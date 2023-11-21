#!/bin/bash

cd $(dirname "$0")

if [ "$UID" != "0" ]; then
  echo "Please run as root!"
  exit 1
fi

# install dependencies
if [ "$(which apt)" != "" ] &>/dev/null; then
  source ./bin/deps/deps.apt.sh
elif [ "$(which dnf)" != "" ] &>/dev/null; then
  source ./bin/deps/deps.dnf.sh
elif [ "$(which rpm-ostree)" != "" ] &>/dev/null; then
  source ./bin/deps/deps.rpm-ostree.sh
else
  echo "Warning: Package Manager Unsupported!"
  echo "Cannot install dependencies"
  echo
  echo "Supported Package Managers:"
  echo "  [apt | dnf | rpm-ostree]"
  echo
  echo "You will need to manually install the following packages:"
  echo "  [crontab | cron | cronie] [inotify-tools] [clamav] [clamav-daemon | clamd clamav-update]"
  echo

  if [ "$1" = "--force" ]; then
    echo "Continuing to install aspiesoft-clamav-scanner..."
  else
    echo "To run the installer anyway, add '--force' to the command"
    exit 1
  fi
fi

systemctl stop clamav-freshclam
freshclam
systemctl enable clamav-freshclam --now
freshclam

# add quarantine folder
if ! [ -d "/VirusScan/quarantine" ]; then
  sudo mkdir -p /VirusScan/quarantine
  sudo chmod 0664 /VirusScan
  sudo chmod 2660 /VirusScan/quarantine
  sudo chmod -R 2660 /VirusScan/quarantine
fi

# fix clamav permissions
if grep -R "^ScanOnAccess " "/etc/clamd.d/scan.conf"; then
  sudo sed -r -i 's/^ScanOnAccess (.*)$/ScanOnAccess yes/m' "/etc/clamd.d/scan.conf"
else
  echo 'ScanOnAccess yes' | sudo tee -a "/etc/clamd.d/scan.conf"
fi

if grep -R "^OnAccessMountPath " "/etc/clamd.d/scan.conf"; then
  sudo sed -r -i 's#^OnAccessMountPath (.*)$#OnAccessMountPath /#m' "/etc/clamd.d/scan.conf"
else
  echo 'OnAccessMountPath /' | sudo tee -a "/etc/clamd.d/scan.conf"
fi

if grep -R "^OnAccessPrevention " "/etc/clamd.d/scan.conf"; then
  sudo sed -r -i 's/^OnAccessPrevention (.*)$/OnAccessPrevention no/m' "/etc/clamd.d/scan.conf"
else
  echo 'OnAccessPrevention no' | sudo tee -a "/etc/clamd.d/scan.conf"
fi

if grep -R "^OnAccessExtraScanning " "/etc/clamd.d/scan.conf"; then
  sudo sed -r -i 's/^OnAccessExtraScanning (.*)$/OnAccessExtraScanning yes/m' "/etc/clamd.d/scan.conf"
else
  echo 'OnAccessExtraScanning yes' | sudo tee -a "/etc/clamd.d/scan.conf"
fi

if grep -R "^OnAccessExcludeUID " "/etc/clamd.d/scan.conf"; then
  sudo sed -r -i 's/^OnAccessExcludeUID (.*)$/OnAccessExcludeUID 0/m' "/etc/clamd.d/scan.conf"
else
  echo 'OnAccessExcludeUID 0' | sudo tee -a "/etc/clamd.d/scan.conf"
fi

if grep -R "^User " "/etc/clamd.d/scan.conf"; then
  sudo sed -r -i 's/^User (.*)$/User root/m' "/etc/clamd.d/scan.conf"
else
  echo 'User root' | sudo tee -a "/etc/clamd.d/scan.conf"
fi

# install aspiesoft clamav download scanner
mkdir -p /etc/aspiesoft-clamav-scanner
cp -rf ./bin/* /etc/aspiesoft-clamav-scanner/
cp -f ./uninstall.sh /etc/aspiesoft-clamav-scanner/
rm -rf /etc/aspiesoft-clamav-scanner/deps

ln -s /etc/aspiesoft-clamav-scanner/aspiesoft-clamav-download-scanner.service /etc/systemd/system/aspiesoft-clamav-download-scanner.service
ln -s /etc/aspiesoft-clamav-scanner/avscan /usr/local/bin/avscan

cd /etc/aspiesoft-clamav-scanner
go build &>/dev/null

systemctl daemon-reload
systemctl enable aspiesoft-clamav-download-scanner.service --now
