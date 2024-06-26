#!/bin/bash

set -ex

if [ "$1" != "--vendor" ]; then
    echo "No vendor provided"
    exit 1
fi
vendor=$2

function get_device_driver() {
    local dev_driver=$(readlink $driver_path)
    echo "${dev_driver##*/}"
}

# find the PCI address of the device by vendor_id:product_id
pci_address=(`lspci -D -d ${vendor}`)
pci_address="${pci_address[0]}"
dev_sysfs_path="/sys/bus/pci/devices/$pci_address"

if [[ ! -d $dev_sysfs_path ]]; then
    echo "Error: PCI address ${pci_address} does not exist!" 1>&2
    exit 1
fi

if [[ ! -d "$dev_sysfs_path/iommu/" ]]; then
    echo "Error: No vIOMMU found in the VM" 1>&2
    exit 1
fi

# set device driver path
driver_path="${dev_sysfs_path}/driver"
driver_override="${dev_sysfs_path}/driver_override"

# load the vfio-pci module
modprobe -i vfio-pci


driver=$(get_device_driver)

if [[ "$driver" != "vfio-pci" ]]; then

    # unbind from the original device driver
    echo ${pci_address} > "${driver_path}/unbind"
    # bind the device to vfio-pci driver
    echo "vfio-pci" > ${driver_override}
    echo $pci_address > /sys/bus/pci/drivers/vfio-pci/bind
fi

# The device should now be using the vfio-pci driver
new_driver=$(get_device_driver)
if [[ $new_driver != "vfio-pci" ]]; then
    echo "Error: Failed to bind to vfio-pci driver" 1>&2
    exit 1
fi
