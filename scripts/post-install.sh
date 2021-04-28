#!/bin/bash

DATA_DIR=/var/lib/thola

if [ ! -f $DATA_DIR/config.yaml ]; then
  echo "loglevel: info" > $DATA_DIR/config.yaml
fi

systemctl enable thola
systemctl restart thola
