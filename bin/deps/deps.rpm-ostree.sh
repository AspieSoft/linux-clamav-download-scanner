#!/bin/bash

rpm-ostree install -y --allow-inactive cronie
rpm-ostree install -y --allow-inactive inotify-tools
rpm-ostree install -y --allow-inactive clamav
rpm-ostree install -y --allow-inactive clamd
rpm-ostree install -y --allow-inactive clamav-update
