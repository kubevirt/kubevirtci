#!/bin/bash

set -ex

systemctl stop origin-node.service
rm -rf /etc/origin/ /etc/etcd/ /var/lib/origin /var/lib/etcd/

containers="$(docker ps -q)"
if [ -n "$containers" ]; then
  docker stop $containers
fi

containers="$(docker ps -q -a)"
if [ -n "$containers" ]; then
  docker rm -f $containers
fi
