# Use kubevirtci with podman instead of docker

Install podman 3.1+, then run it in docker compatible mode:

## Rootless podman

### Setup (Fedora)

Enable the user podman socket:

```bash
systemctl --user enable --now podman.socket
```

Load required kernel modules now and during future boots:

```bash
sudo tee /etc/modules-load.d/kubevirtci.conf <<EOF
# Kernel modules required for KubeVirt cluster-up
ip_tables
iptable_nat
iptable_filter
ip6_tables
ip6table_nat
ip6table_filter
kvm
EOF
sudo systemctl restart systemd-modules-load
```

Configure the podman socket path for your future shells:

```bash
mkdir -p ~/.bashrc.d
tee ~/.bashrc.d/kubevirtci <<EOF
export KUBEVIRTCI_PODMAN_SOCKET=\$XDG_RUNTIME_DIR/podman/podman.sock
EOF
```

## Rootful podman

In order to use rootful podman by a non root user, we will need to bind podman
to a socket, accessible by the user (as docker does).

Assuming the user is in `wheel` group please do the following (one time):

As root, create a Drop-In file `/etc/systemd/system/podman.socket.d/10-socketgroup.conf`
with the following content:
```
[Socket]
SocketGroup=wheel
ExecStartPost=/usr/bin/chmod 755 /run/podman
```

The 1st line is needed in order to create the socket accessible by the `wheel` group.
2nd line because systemd-tmpfiles recreates the folder as root:root without group reading rights.

Stop `podman.socket` if it is running,
reload the daemon `systemctl daemon-reload` since we changed the systemd settings
and restart it again `systemctl enable --now podman.socket`

As the user add the following to your ~/.bashrc
```
alias podman="podman --remote"
export CONTAINER_HOST=unix:///run/podman/podman.sock
```

Validate it by running `podman run hello-world` as the non root user
and see that as root `podman ps -a` shows the same exited container (or vice versa).

In case you wish to use a custom socket path, change the values of `CONTAINER_HOST`
and `KUBEVIRTCI_PODMAN_SOCKET` accordingly,
i.e `export KUBEVIRTCI_PODMAN_SOCKET="${XDG_RUNTIME_DIR}/podman/podman.sock"`

Tested on fedora 35.

## Resource Adjustments

When working with Podman, you might encounter PID resource constraints. To resolve this issue:

1. Locate and edit your `containers.conf` file (typically in `/usr/share/containers`)
2. Add or modify the PID limit setting:
    ```toml
    [containers]
    # Configure the process ID (PID) limit for containers
    # Options:
    #   Numeric value: Sets specific PID limit (e.g., 2048)
    #   -1: Removes PID limitations entirely
    pids_limit = 2048
    ```
