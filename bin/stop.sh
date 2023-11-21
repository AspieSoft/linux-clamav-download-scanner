#!/bin/bash

cd $(dirname "$0")

rm -f ./running.tmp

sleep 1

killall linux-clamav-download-scanner
