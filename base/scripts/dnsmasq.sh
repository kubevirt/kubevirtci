#!/bin/bash

set -e

NUM_NODES=${NUM_NODES-1}

ip link add br0 type bridge
ip link set dev br0 up
ip addr add dev br0 192.168.66.02/24

for i in $(seq 1 ${NUM_NODES}); do
  n="$(printf "%02d" ${i})"
  ip tuntap add dev tap${n} mode tap user $(whoami)
  ip link set tap${n} master br0
  ip link set dev tap${n} up
  DHCP_HOSTS="${DHCP_HOSTS} --dhcp-host=52:55:00:d1:55:${n},192.168.66.1${n},node${n},infinite"
done

# Make sure that all VMs can reach the internet
iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
iptables -A FORWARD -i br0 -o eth0 -j ACCEPT

# Tell the first node that it can start provisioning
if [ -n "${NEXT_HOST}" ]; then
  touch /shared/${NEXT_HOST}.start
fi

exec dnsmasq -d ${DHCP_HOSTS} --dhcp-range=192.168.66.10,192.168.66.200,infinite
