#!/bin/bash

if [ "$1" != "" ]; then
  DBUS_SESSION_BUS_ADDRESS="$1" notify-send -i "$2" -t 3 "Started Scanning File" "$3"
else
  notify-send -i "$2" -t 3 "$3" "$4"
fi
