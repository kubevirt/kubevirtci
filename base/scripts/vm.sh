#!/bin/bash

set -e

PROVISION=false
MEMORY=3096M
CPU=2
QEMU_ARGS=""

while true; do
  case "$1" in
    -m | --memory ) MEMORY="$2"; shift 2 ;;
    -c | --cpu ) CPU="$2"; shift 2 ;;
    -q | --qemu-args ) QEMU_ARGS="$2"; shift 2 ;;
    -- ) shift; break ;;
    * ) break ;;
  esac
done

function calc_next_disk {
  last="$(ls -t disk* | head -1 | sed -e 's/disk//' -e 's/.qcow2//')"
  last="${last:-00}"
  next=$((last+1))
  next=$(printf "disk%02d.qcow2" $next)
  if [ "$last" = "00" ]; then
    last="box.qcow2"
  else
    last=$(printf "disk%02d.qcow2" $last)
  fi
}

NODE_NUM=${NODE_NUM-1}
n="$(printf "%02d" ${NODE_NUM})"

cat >/usr/local/bin/ssh.sh <<EOL
#!/bin/bash
set -e
dockerize -wait tcp://192.168.66.1${n}:22 -timeout 300s
ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no vagrant@192.168.66.1${n} -i vagrant.key -p 22 -q \$@
EOL
chmod u+x /usr/local/bin/ssh.sh

sleep 0.1
until ip link show tap${n}; do
  echo "Waiting for tap${n} to become ready"
  sleep 0.1
done

# Route SSH
iptables -t nat -A POSTROUTING --out-interface br0 -j MASQUERADE
iptables -A FORWARD --in-interface eth0 -j ACCEPT
iptables -t nat -A PREROUTING -p tcp -i eth0 -m tcp --dport 22${n} -j DNAT --to-destination 192.168.66.1${n}:22

# Route 6443 and 8443 for first node
if [ "$n" = "01" ] ; then
  iptables -t nat -A PREROUTING -p tcp -i eth0 -m tcp --dport 6443 -j DNAT --to-destination 192.168.66.1${n}:6443
  iptables -t nat -A PREROUTING -p tcp -i eth0 -m tcp --dport 8443 -j DNAT --to-destination 192.168.66.1${n}:8443
fi

# For backward compatibility, so that we can just copy over the newer files
if [ -f provisioned.qcow2 ]; then
  ln -sf provisioned.qcow2 disk01.qcow2
fi

calc_next_disk

echo "Creating disk \"${next} backed by ${last}\"."
qemu-img create -f qcow2 -o backing_file=${last} ${next}

echo ""
echo "SSH will be available on container port 22${n}."
echo "VNC will be available on container port 59${n}."
echo "VM MAC in the guest network will be 52:55:00:d1:55:${n}"
echo "VM IP in the guest network will be 192.168.66.1${n}"
echo "VM hostname will be node${n}"

exec qemu-kvm -drive format=qcow2,file=${next}  \
  -device e1000,netdev=network0,mac=52:55:00:d1:55:${n} \
  -netdev tap,id=network0,ifname=tap${n},script=no,downscript=no \
  -vnc :${n} -enable-kvm -cpu host -m ${MEMORY} -smp ${CPU} ${QEMU_ARGS}
