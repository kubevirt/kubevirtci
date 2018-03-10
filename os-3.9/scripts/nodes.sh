#!/bin/bash

set -e

# disable all master components, the machine should be a node
systemctl stop origin-master-api
systemctl disable origin-master-api
systemctl stop origin-master-controllers
systemctl disable origin-master-controllers

# make sure that dns works properly, openshift installer messes something up
dhclient
