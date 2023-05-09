# Customize ephemeral container-disk images for KubeVirt VM's

This tool is targeted for KubeVirt developers, users or anyone who
would like to create KubeVirt VM's with customized container-disk,
without the hassle of installing and customizing the OS manually.

Use cases:
- VM's for Demos
- Pre-configured Kubevirt VM image for tests, for example:
    - Image with qemu-guest-agent installed and configured.
    - Image with sriov drivers for sriov-lane images.
    - Image with DPDK package for testing DPDK applications on KubeVirt.


## Prerequisites

The following RPM packages need to be present on your machine:
- cloud-utils
- docker-ce
- libguestfs
- libguestfs-tools-c
- libvirt
- qemu-img
- virt-install

To cross build for the Arm64 image on x86_64 machines, the following RPM needs to be installed:
- qemu-system-aarch64

## Quickstart: Build and publish an existing containerdisk

Choose one of the configuration directories in this folder which you want to
build. In this case the `example` directory is chosen.

To build the `example` directory, run

```
create-containerdisk.sh example
```

then publish it to a registry. In this case a local registry:

```
publish-containerdisk.sh example localhost:1234/myimage:mytag
```

To build for aarch64(arm64), you need to set the following environment
variable.
```
export ARCHITECTURE=aarch64
```

To build the Virtual Machine without console output, only need to set the
environment variable `CONSOLE`. This is useful when the build script is
run in a CICD pipeline.
```
export CONSOLE=no
```
Then follow the previous step.
## Create a new containerdisk

Every directory contains build instructions for virt-customize. The build
instructions are distributed between the following three files:

```
$ ls -1 example/
cloud-config		#cloud-init configuration for virt-customize
image-url		#download URL of the base image
image-url-aarch64	#download URL of the base image for aarch64
os-variant		#operating system variant (for example fedora32)
```

To create a completely new containerdisk, best copy the `example` folder and
customize the three files.

## Push the new image to local cluster registry:
```bash
# From kubevirtci / kubevirt directory
$ ./create-containerdisk.sh example
$ ./publish-containerdisk.sh example "localhost:$(./cluster-up/cli.sh ports registry | tr -d '\r')"
```

### Create Kubevirt VM's with the new image

In the VMI / VM yaml file, change `spec.volumes[].containerDisk.image` to the new image path.

It is possible to use an image from local cluster registry.

```bash
$ kubectl apply -f <VMI yaml file>
$ kubectl wait --for=condition=AgentConnected vmi $VMI_NAME --timeout 5m
$ virtctl console testvm1
```

### Build and publish multi-arch images
The multi-arch publish does not support building alpine-cloud-init because the [alpine-make-vm-image](https://raw.githubusercontent.com/alpinelinux/alpine-make-vm-image/master/alpine-make-vm-image) project does not support building Arm64 images.
The `publish-multiarch-containerdisk.sh` script now supports building Arm64 and x86_64 images.
The script primarily performs the following tasks:
1. Use `create-containerdisk.sh` to build images.
2. Upload the resulting images to a specific registry.
3. Upload multi-arch image manifest.

```
./publish-multiarch-containerdisk.sh -h
    Usage:
        ./publish_multiarch_image.sh [OPTIONS] BUILD_TARGET REGISTRY REGISTRY_ORG

    Build and publish multiarch infra images.

    OPTIONS
        -n  (native build) Only build image for host CPU Arch.
        -h  Show this help message and exit.
        -b  Only build the image and exit. Do not publish the built image.

# build and push multi-arch example image
./ publish-multiarch-containerdisk.sh example myregistry registry_org

# The script will do following things:
# 1. Build both Arm64 and x86_64 example image.
# 2. Generate a tag based on the current time.
# 3. Push the registry_org/myregistry/example:tag-aarch64 and registry_org/myregistry/example:tag-x86_64.
# 4. Generate a multi-arch manifest for the image, registry_org/myregistry/example:tag.
```

To use your own tag for the image, you need to set the environment variable `KUBEVIRTCI_TAG`.
```
export KUBEVIRTCI_TAG=mytag
```
