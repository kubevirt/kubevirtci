#!/bin/bash

set -e

NUM_NODES=${NUM_NODES-1}
NUM_SECONDARY_NICS=${NUM_SECONDARY_NICS:-0}

ip link add br0 type bridge
ip link set dev br0 up
ip addr add dev br0 192.168.66.02/24

# Create secondary networks
for snet in $(seq 1 ${NUM_SECONDARY_NICS}); do
  ip link add br${snet} type bridge
  ip link set dev br${snet} up
done

for i in $(seq 1 ${NUM_NODES}); do
  n="$(printf "%02d" ${i})"
  ip tuntap add dev tap${n} mode tap user $(whoami)
  ip link set tap${n} master br0
  ip link set dev tap${n} up
  DHCP_HOSTS="${DHCP_HOSTS} --dhcp-host=52:55:00:d1:55:${n},192.168.66.1${n},node${n},infinite"
  for s in $(seq 1 ${NUM_SECONDARY_NICS}); do
    tap_name=stap$(($i - 1))-$(($s - 1))
    ip tuntap add dev $tap_name mode tap user $(whoami)
    ip link set $tap_name master br${s}
    ip link set dev $tap_name up
  done
done

# Make sure that all VMs can reach the internet
iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
iptables -A FORWARD -i br0 -o eth0 -j ACCEPT

exec dnsmasq -d ${DHCP_HOSTS} --dhcp-range=192.168.66.10,192.168.66.200,infinite
