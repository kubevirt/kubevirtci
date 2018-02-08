set -e

i=${NODE_INDEX-1}

sleep 0.1
until ip link show tap${i}; do
  echo "Waiting for tap${i} to become ready"
  sleep 0.1
done

n="$(printf "%02d" ${i})"

if [ "$i" = "1" ]; then
  iptables -t nat -A POSTROUTING --out-interface br0 -j MASQUERADE
  iptables -A FORWARD --in-interface eth0 -j ACCEPT
  iptables -t nat -A PREROUTING -p tcp -i eth0 -m tcp --dport 22 -j DNAT --to-destination 192.168.66.1${n}:22
fi

# Create a transient disk, so that the container runtime does not have to copy the whole file on writes
qemu-img create -f qcow2 -o backing_file=box.qcow2 disk.qcow2

exec qemu-kvm -drive format=qcow2,file=disk.qcow2 \
  -device e1000,netdev=network0,mac=52:55:00:d1:55:${n} \
  -netdev tap,id=network0,ifname=tap${i},script=no,downscript=no \
  -vnc :1 -enable-kvm -cpu host -m 3096M
