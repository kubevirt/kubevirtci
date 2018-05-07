# Distroless QEMU

A static build of qemu, stuffed into a distroless docker container. Includes
all usual qemu utilities like `qemu-img` or `qemu-nbd` and the
`qemu-system-x86_64`.

## Build the QEMU container

Build the container:

```bash
bazel run //:qemu_image
```

Let's have a look at the container size:

```bash
$  docker history bazel:qemu_image
IMAGE               CREATED             CREATED BY          SIZE                COMMENT
262453680ed8        48 years ago        bazel build ...     36.3 MB
```

**36.3 MB**.

## Run a cirros image

Create and run a test container which contains an extra layer on the qemu image
with a cirros qcow2 image:

```bash
bazel run //:qemu_image_test
docker run -it --rm -p 127.0.0.1:5901:5901 bazel:qemu_image_test -vnc 0.0.0.0:01 -drive format=qcow2,file=/disk/cirros.img -vga std
```

Connect via VNC:

```bash
remote-viewer vnc://localhost:5901
```
