#! /bin/bash

set -e

function configure_calico(){
    cat <<EOF >/etc/NetworkManager/conf.d/calico.conf
[keyfile]
unmanaged-devices=interface-name:cali*;interface-name:tunl*
EOF

    sysctl -w net.netfilter.nf_conntrack_max=1000000
    echo "net.netfilter.nf_conntrack_max=1000000" >> /etc/sysctl.conf

    systemctl restart NetworkManager
}
