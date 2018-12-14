#!/usr/bin/env bash

IMAGES="k8s/1.10.4 k8s/1.11.0 k8s-multus/1.10.4 k8s-multus/1.11.1 os-3.10 os-3.10-crio os-3.10-multus os-3.11-multus-sriov"
LOG=`pwd`/log.txt

echo
echo "Provisioning and publishing these images, this will take a while:"
echo $IMAGES
echo
echo "All logs will go to $LOG, here you will only see the final shas for each image!"
echo
read -p "Do you want to continue? [y/n] " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    echo
    echo "cancelled"
    exit 0
fi

echo "" > $LOG

for IMAGE in $IMAGES
do
    echo
    date | tee -a $LOG
    echo "building $IMAGE" | tee -a $LOG
    {
        (cd $IMAGE && ./provision.sh && ./publish.sh)
    } 2>&1 | tee -a $LOG | grep -i -e 'The push refers to' -e 'latest: digest:'
    printf "\n\n\n" >> $LOG
done

date | tee -a $LOG
echo "all done" | tee -a $LOG
