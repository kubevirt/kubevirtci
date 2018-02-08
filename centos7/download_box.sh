#!/bin/bash

set -e
set -o pipefail

curl http://cloud.centos.org/centos/7/vagrant/x86_64/images/CentOS-7-x86_64-Vagrant-${1}.Libvirt.box | tar -zxvf - box.img
