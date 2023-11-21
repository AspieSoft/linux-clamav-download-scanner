#!/bin/bash

cd $(dirname "$0")

# add cron job for module updates
if ! [[ $(crontab -l) == *"# aspiesoft-clamav-scan-update"* ]] ; then
  crontab -l | { cat; echo '0 2 * * * sudo nice -n 10 /etc/aspiesoft-clamav-scanner/update.sh # aspiesoft-clamav-scan-update'; } | crontab -
fi

# add cron job for full system scan
if ! [[ $(crontab -l) == *"# aspiesoft-clamav-scan"* ]] ; then
  crontab -l | { cat; echo '0 2 * * * sudo nice -n 15 clamscan -r --bell --move="/VirusScan/quarantine" --exclude-dir="/VirusScan/quarantine" --exclude-dir="/home/$USER/.clamtk/viruses" --exclude-dir="smb4k" --exclude-dir="/run/user/$USER/gvfs" --exclude-dir="/home/$USER/.gvfs" --exclude-dir=".thunderbird" --exclude-dir=".mozilla-thunderbird" --exclude-dir=".evolution" --exclude-dir="Mail" --exclude-dir="kmail" --exclude-dir="^/sys" / # aspiesoft-clamav-scan'; } | crontab -
fi

# add this file so we can easily stop the loop from running in the stop.sh script
echo "running" > ./running.tmp

# scan downloads and new files
# by default, this will only scan common files in the current users home directory, and extension files (browser extensions, etc)
while true; do
  GOMEMLIMIT="$((1024 * 1024 * 200))" /etc/aspiesoft-clamav-scanner/linux-clamav-download-scanner
  sleep 1

  if ! [ -f ./running.tmp ]; then
    exit
  fi
done
