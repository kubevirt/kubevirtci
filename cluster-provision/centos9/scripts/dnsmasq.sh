#!/bin/bash

set -ex

NUM_NODES=${NUM_NODES-1}
NUM_SECONDARY_NICS=0
SEQ_START=1

ip link add br0 type bridge
echo 0 > /proc/sys/net/ipv6/conf/br0/disable_ipv6
echo 1 > /proc/sys/net/ipv6/conf/all/forwarding
ip link set dev br0 up
ip addr add dev br0 192.168.66.02/24
ip -6 addr add fd00::1/64 dev br0

# Create secondary networks
for snet in $(seq 1 ${NUM_SECONDARY_NICS}); do
  ip link add br${snet} type bridge
  ip link set dev br${snet} up
done

# if the number is one then do the normal thing, if its higher than one then do the manual ip thingy
i=1
while [ $i -le ${NUM_NODES} ]; do
  if [ ${NUM_NODES} -gt 1 ]; then
    ip tuntap add dev tap101 mode tap user $(whoami)
    ip link set tap101 master br0
    ip link set dev tap101 up
    ip addr add 192.168.66.110/24 dev tap101
    ip -6 addr add fd00::110 dev tap101
    iptables -t nat -A PREROUTING -p tcp -i eth0 -m tcp --dport 6443 -j DNAT --to-destination 192.168.66.110:6443
  fi

  n="$(printf "%02d" ${i})"
  ip tuntap add dev tap${n} mode tap user $(whoami)
  ip link set tap${n} master br0
  ip link set dev tap${n} up
  DHCP_HOSTS="${DHCP_HOSTS} --dhcp-host=52:55:00:d1:55:${n},192.168.66.1${n},[fd00::1${n}],node${n},infinite"

  for s in $(seq 1 ${NUM_SECONDARY_NICS}); do
    tap_name=stap$(($i - 1))-$(($s - 1))
    ip tuntap add dev $tap_name mode tap user $(whoami)
    ip link set $tap_name master br${s}
    ip link set dev $tap_name up
  done

  ((i++))  # Increment for the main loop
done

# Make sure that all VMs can reach the internet
iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
iptables -A FORWARD -i br0 -o eth0 -j ACCEPT
ip6tables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
ip6tables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
ip6tables -A FORWARD -i br0 -o eth0 -j ACCEPT

exec dnsmasq --interface=br0 --enable-ra --dhcp-option=option6:dns-server,[::] -d ${DHCP_HOSTS} --dhcp-range=192.168.66.10,192.168.66.200,infinite --dhcp-range=::10,::200,constructor:br0,static