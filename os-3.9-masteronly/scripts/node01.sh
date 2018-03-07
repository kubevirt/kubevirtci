#!/bin/bash

set +e

/usr/local/bin/oc get nodes
while [ $? -ne 0 ]; do
    sleep 5
    /usr/local/bin/oc get nodes
done
