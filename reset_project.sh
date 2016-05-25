#!/bin/bash

#
# Reset everything in vagrant box
#

# remove all containers
CONTAINER_LIST=$(docker ps -a -q)
if [ ! -z "$CONTAINER_LIST" ]; then
    docker rm -f ${CONTAINER_LIST}
fi

# remove all images
IMAGES_LIST=$(docker images -q)
if [ ! -z "$IMAGES_LIST" ]; then
    docker rmi ${IMAGES_LIST}
fi

# remove deploy log
if [ -d "/home/vagrant/.zdd" ]; then
    rm -rf "/home/vagrant/.zdd"
fi
