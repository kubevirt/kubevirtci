# Use kubevirtci with podman instead of docker

Install podman 3.1+, then run it in docker compatible mode:

```
systemctl start --user podman.socket
```

After this command, you can use kubevirtci with the typical `make cluster-*`
commands.

This will use `fuse-overlayfs` as storage layer. If the performance is not
satisfactory, consider running podman as root to use plain `overlayfs2`:

```
mkdir -p $XDG_RUNTIME_DIR/podman
sudo podman system service -t 0 unix:///$XDG_RUNTIME_DIR/podman/podman.sock
sudo chown $USER $XDG_RUNTIME_DIR/podman/podman.sock
```

Note that `podman system service` will keep running in the foreground so the
current terminal must be kept open and the last command must be executed in a
new terminal.
