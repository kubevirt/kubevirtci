#!/bin/bash
set -e

# Setup bridges for secondary network interfaces (eth1, eth2, eth3, ...)
# For each ethX interface where X >= 1, create a corresponding brX bridge
# Uses NetworkManager (nmcli) to create persistent bridge configurations

for iface in $(ls /sys/class/net/ 2>/dev/null | grep -E '^eth[1-9]$|^eth[0-9]{2,}$'); do
    bridge_name="br${iface#eth}"

    echo "Setting up bridge ${bridge_name} for interface ${iface}"

    # Delete existing connections for the interface if any (to ensure clean state)
    nmcli connection delete ${iface} 2>/dev/null || true
    nmcli connection delete ${bridge_name} 2>/dev/null || true

    # Create bridge connection with autoconnect enabled
    nmcli connection add type bridge ifname ${bridge_name} con-name ${bridge_name} \
        connection.autoconnect yes \
        connection.autoconnect-priority 100

    # Add ethernet interface as bridge slave with autoconnect
    nmcli connection add type ethernet ifname ${iface} con-name ${iface} master ${bridge_name} \
        connection.autoconnect yes \
        connection.autoconnect-priority 100

    # Bring up the bridge (which will also bring up the slave)
    nmcli connection up ${bridge_name}

    # Verify the bridge was created successfully
    if ip link show ${bridge_name} &>/dev/null; then
        echo "Bridge ${bridge_name} created and verified successfully"
        # Show the bridge details
        ip link show ${bridge_name}
        ip link show ${iface} | grep -q "master ${bridge_name}" && echo "  ✓ ${iface} enslaved to ${bridge_name}"
    else
        echo "ERROR: Bridge ${bridge_name} creation failed!"
        exit 1
    fi
done

echo "All secondary interface bridges configured"

# Final verification
echo "Verifying all bridges..."
ip link show | grep -E "^[0-9]+: br[0-9]+"
