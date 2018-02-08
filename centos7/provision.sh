#!/bin/bash

i=${NODE_INDEX-1}

ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no vagrant@192.168.66.1${i} -i vagrant.key < provision_once.sh
