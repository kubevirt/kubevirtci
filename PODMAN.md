# Use kubevirtci with podman instead of docker

Install podman 3.1+, then run it in docker compatible mode:

```
podman system service -t 0 unix:///${HOME}/podman.sock
```

After this command, you can use kubevirtci with the typical `make cluster-*`
commands.

This will use `fuse-overlayfs` as storage layer. If the performance is not
satisfactory, consider running podman as root to use plain `overlayfs2`:

```
sudo podman system service -t 0 unix:///${HOME}/podman.sock
```
