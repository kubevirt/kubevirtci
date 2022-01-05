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
