# Use kubevirtci with podman instead of docker

Install podman 3.1+, then run it in docker compatible mode:

## Rootless podman

```
systemctl start --user podman.socket
```

Currently rootless podman is **not** working with the `make cluster-sync`
command, essentially because incoming traffic is coming from the loopback device
instead of eth0.

The current rules - [ssh](https://github.com/kubevirt/kubevirtci/blob/962d90cead28fc2aadcc07388b18d2479b2b6714/cluster-provision/centos8/scripts/vm.sh#L73), [restricted ports](https://github.com/kubevirt/kubevirtci/blob/962d90cead28fc2aadcc07388b18d2479b2b6714/cluster-provision/centos8/scripts/vm.sh#L83) - allow `make cluster-up` to run successfully, but
unfortunately they break the cluster's network connectivity in a subtle way:
image pulling fails because outgoing traffic to ports 22 6443 8443 80 443 30007
30008 31001 is redirected to the VM in the respective node container (i.e.
itself) instead of going to the specified host (e.g. quay.io).

This will use `fuse-overlayfs` as storage layer. If the performance is not
satisfactory, consider running podman as root to use plain `overlayfs2`:

## Rootful podman

```
mkdir -p $XDG_RUNTIME_DIR/podman
sudo podman system service -t 0 unix:///$XDG_RUNTIME_DIR/podman/podman.sock
sudo chown $USER $XDG_RUNTIME_DIR/podman/podman.sock
```

After this command, you can use kubevirtci with the typical `make cluster-*`
commands.

Note that `podman system service` will keep running in the foreground so the
current terminal must be kept open and the last command must be executed in a
new terminal.

