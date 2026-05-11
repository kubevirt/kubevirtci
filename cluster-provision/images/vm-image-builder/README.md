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
- podman
- libguestfs
- libguestfs-tools-c
- libvirt
- qemu-img
- virt-install

To cross build for the Arm64 image on AMD64 machines, the following RPM needs to be installed:
- qemu-system-aarch64

To cross build for the s390x image on AMD64 machines, the following RPM needs to be installed:
- qemu-system-s390x

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

To build for arm64(aarch64), you need to set the following environment
variable.
```
export ARCHITECTURE=arm64
```

To build for s390x, you need to set the following environment
variable.
```
export ARCHITECTURE=s390x
``` 

To build the Virtual Machine without console output, only need to set the
environment variable `CONSOLE`. This is useful when the build script is
run in a CICD pipeline.
```
export CONSOLE=false
```
Then follow the previous step.
## Create a new containerdisk

Every directory contains build instructions for virt-customize. The build
instructions are distributed between the following three files:

```
$ ls -l example/
cloud-config    # cloud-init configuration for virt-customize
image-url       # download URL of the base image
image-url-arm64 # download URL of the base image for Arm64
image-url-s390x # download URL of the base image for s390x
os-variant      # operating system variant (for example fedora32)
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
The multi-arch publish does not support building alpine-cloud-init because the [alpine-make-vm-image](https://raw.githubusercontent.com/alpinelinux/alpine-make-vm-image/master/alpine-make-vm-image) project does not support building Arm64 and s390x images.
The `publish-multiarch-containerdisk.sh` script now supports building Arm64, AMD64 and s390x images.
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
# 1. Build the Arm64,AMD64 and s390x example image.
# 2. Generate a tag based on the current time.
# 3. Push the registry_org/myregistry/example:tag-arm64,registry_org/myregistry/example:tag-amd64 and registry_org/myregistry/example:tag-s390x
# 4. Generate a multi-arch manifest for the image, registry_org/myregistry/example:tag.
```

To use your own tag for the image, you need to set the environment variable `KUBEVIRTCI_TAG`.
```
export KUBEVIRTCI_TAG=mytag
```

## End-to-end: updating fedora-with-test-tooling in KubeVirt

The `fedora-with-test-tooling` container disk is used by KubeVirt e2e tests. Unlike
the alpine image and the k8s provider images, it has no automated CI job for
publishing. The full update cycle is manual and spans two repositories.

### 1. Make changes (kubevirtci)

Edit the files under `fedora-with-test-tooling/` (e.g. `cloud-config` to add
packages or configuration).

### 2. Build locally (kubevirtci)

Build the image for your host architecture to verify the changes:

```bash
cd cluster-provision/images/vm-image-builder
./create-containerdisk.sh fedora-with-test-tooling
```

This downloads the base Fedora cloud image, boots a VM with `cloud-config` via
`virt-install`, runs `virt-sysprep`, and wraps the result in a `FROM scratch`
container image.

Prerequisites: see the [Prerequisites](#prerequisites) section above.

### 3. Publish to quay.io (kubevirtci)

Build and push for all architectures (amd64, arm64, s390x):

```bash
cd cluster-provision/images/vm-image-builder
./publish-multiarch-containerdisk.sh fedora-with-test-tooling quay.io kubevirtci
```

This pushes per-arch images and a multi-arch manifest to
`quay.io/kubevirtci/fedora-with-test-tooling:<tag>`.

You need write access to the `kubevirtci` organization on quay.io.

### 4. Get the new image digests

After publishing, retrieve the per-arch digests:

```bash
skopeo inspect --raw docker://quay.io/kubevirtci/fedora-with-test-tooling:<tag> | \
  jq '.manifests[] | {platform: .platform, digest: .digest}'
```

### 5. Update the digest in kubevirt (kubevirt repo)

In the `kubevirt/kubevirt` repository, update the `oci_pull` entries in the
`WORKSPACE` file with the new digests:

```python
oci_pull(
    name = "fedora_with_test_tooling",
    digest = "sha256:<new-amd64-digest>",
    image = "quay.io/kubevirtci/fedora-with-test-tooling",
)

oci_pull(
    name = "fedora_with_test_tooling_aarch64",
    digest = "sha256:<new-arm64-digest>",
    image = "quay.io/kubevirtci/fedora-with-test-tooling",
)

oci_pull(
    name = "fedora_with_test_tooling_s390x",
    digest = "sha256:<new-s390x-digest>",
    image = "quay.io/kubevirtci/fedora-with-test-tooling",
)
```

### 6. Verify

In the kubevirt repo, run the e2e tests that use
`fedora-with-test-tooling-container-disk` to confirm the updated image works
correctly.

### Summary of the image flow

```
kubevirtci                                         kubevirt
────────                                           ───────
fedora-with-test-tooling/cloud-config
        │
        ▼
create-containerdisk.sh
  ├─ downloads base Fedora cloud image
  ├─ boots VM with cloud-init (virt-install)
  ├─ virt-sysprep + compress
  └─ wraps qcow2 in FROM scratch container
        │
        ▼
publish-multiarch-containerdisk.sh
  ├─ builds amd64, arm64, s390x
  ├─ pushes per-arch images
  └─ pushes multi-arch manifest
        │
        └──► quay.io/kubevirtci/fedora-with-test-tooling:<tag>
                                                        │
                                                        ▼
                                              WORKSPACE (oci_pull by digest)
                                                        │
                                                        ▼
                                              containerimages/BUILD.bazel
                                              (oci_image + oci_push)
                                                        │
                                                        ▼
                                              quay.io/kubevirt/fedora-with-test-tooling-container-disk
                                              (used in e2e test VMI specs)
```
