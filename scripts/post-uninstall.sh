#!/bin/bash

if [[ "$1" != "upgrade" ]]; then
  systemctl stop thola
  systemctl disable thola
  rm -f /lib/systemd/system/thola.service
fi
