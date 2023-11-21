# Linux ClamAV Download Scanner

[![donation link](https://img.shields.io/badge/buy%20me%20a%20coffee-paypal-blue)](https://paypal.me/shaynejrtaylor?country.x=US&locale.x=en_US)

> This module is currently in beta.

Automatically scan your linux home directory when you download something new.

By default, this module only scans common directories (Downloads, Desktop, etc.) and searches for extension directories like your chrome extensions.
I may add additional directories in future updates.
Currently, this module only uses the active users home directory, and does not touch the root directory.

## Installation

```shell script
git clone https://github.com/AspieSoft/linux-clamav-download-scanner.git
sudo ./linux-clamav-download-scanner/install.sh
rm -rf ./linux-clamav-download-scanner
```

## Config

add directories for to auto scan on downloads / new or modified files

```shell script
nano ~/.aspiesoft-clamav-auto-scan
# or
nano ~/.clamav-auto-scan

# list files without the /home/username/
Downloads
.config
```

### To add default directories for all users

```shell script
nano /usr/share/config/aspiesoft-clamav-auto-scan
# or
nano /usr/share/config/clamav-auto-scan

# list files without the /home/username/
Downloads
.config
```

## Uninstall

```shell script
sudo /etc/aspiesoft-clamav-scanner/uninstall.sh
```
