#!/bin/bash

set +e

# Wait until cluster will be up
/usr/local/bin/oc get nodes
while [ $? -ne 0 ]; do
    sleep 5
    /usr/local/bin/oc get nodes
done

# Update DHCP lease, will also update DNS servers
dhclient
