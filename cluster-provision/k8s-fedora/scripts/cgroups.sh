#!/bin/bash
# This script checks if the OS is fedora31,
# and configures the system to work with cgroup v1 if needed.

set -e

TARGET_ID=${TARGET_ID:-"fedora"}
TARGET_ID_VERSION=${TARGET_ID_VERSION:-31}

function get_cgroup_hierarchy {
    kernel_info=$(sudo grubby --info=$(sudo grubby --default-kernel))
    echo $(echo $kernel_info | grep -Po 'systemd.unified_cgroup_hierarchy=\K\d*')
}

function get_os_major_version(){
    echo $(cat /etc/os-release | grep -Po '^VERSION_ID=\K.*')
}

function get_os_id(){
    echo $(cat /etc/os-release | grep -Po '^ID=\K\S*')
}

if [[ $(rpm -q grubby) =~ "not installed" ]]; then
   echo installing grubby 
   yum install -y grubby
fi

os=$(get_os_id)
os_version=$(get_os_major_version)
cgroup_hierarchy=$(get_cgroup_hierarchy)
echo "OS: $os $os_version unified_cgroup_hierarchy: $cgroup_hierarchy"
if [[ $os =~ $TARGET_ID ]]; then    
    if [ $os_version == $TARGET_ID_VERSION ]; then
        # Revert cgroup to v1 in order to work with docker.
        if [[ $cgroup_hierarchy == 0 ]]; then
            echo "cgroup configured properly to work with docker"
        else
            grubby --update-kernel=ALL --args="systemd.unified_cgroup_hierarchy=0"
            reboot
        fi
    fi
fi

