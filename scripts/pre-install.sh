#!/bin/bash

DATA_DIR=/var/lib/thola
USER=thola

if ! id thola &>/dev/null; then
    useradd --system -U -M thola -s /bin/false -d $DATA_DIR
fi

if [ ! -d $DATA_DIR ]; then
    mkdir -p $DATA_DIR
    chown $USER:$USER $DATA_DIR
fi