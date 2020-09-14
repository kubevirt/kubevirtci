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

## Create customized container-disk

In order to create customized  container-disk for KubeVirt VM,
use `create-containderdisk.sh` script.

This script automates the process of customizing VM image and build Kubevirt VM container-disk-image.
Soon there will be used for creating and image automatically using CI.

What this script does:
- Download VM image from given URL

- Customize VM image using `customize-image.sh` script, according to given cloud-config file.
  Keeping customizing image method loosely-coupled, so it will be easy to maintain.
  You can use any image customizing script by exporting `CUSTOMIZE_IMAGE_SCRIPT=<script path>`.

- Building container-disk image using `build-containerdisk.sh`
  according to the doc:
  https://github.com/kubevirt/kubevirt/blob/master/docs/container-register-disks.md

- Store the image in local registry and exports the container image to .tar file, so it will be easier to store or send.

This script expected variables:  
- `VM_IMAGE_URL` - URL to download the image
- `CLOUD_CONFIG_PATH` - cloud-conifg file path used  for customizing the image with cloud-init
- `IMAGE_NAME`, `TAG` - container-disk image tag that will be created tag details `IMAGE_NAME:TAG`
- `OS_VARIANT`- required by customize-image.sh

Notes:
- If no variables passed, the script will create container-disk image according to the files at `example/`.
- Currently, image customization process done by spinning up live VM with cloud-init disk to provision the image with the changes described by the cloud-config file.
- It is necessary to add `sudo shutdown` at the end of the cloud-config`runcmd` block, so the script will not wait for user input.
- You can use your own image customizing script by exporting the path to `CUSTOMIZE_IMAGE_SCRIPT`.

This script requires the following packages:
- cloud-utils
- docker-ce
- libvirt
- libguestfs
- qemu-img

## How to use:
```bash
cd cluster-provision/images/container-disk-images/
# Export image name and tag
$ export IMAGE_NAME="example-fedora"
$ export TAG="32"

# Export VM image URL:
$ export IMAGE_URL=$(cat example/image-url)
$ export OS_VARIANT=fedora31

# Export cloud-config file
$ export CLOUD_CONFIG_PATH="example/cloud-config"

# Run script
$ ./create-containerdisk.sh
...
Successfully tagged example-fedora:32
...
Container image saved as tar file at: example-fedora_build/example-fedora-32.tar
...

# You can send / store the image .tar file:
$ ls example-fedora_build/example-fedora-32.tar
example-fedora_build/example-fedora-32.tar
```

## Publish customized container-disk image

Push the new container-disk image to a registry with `publish-containerdisk.sh` script.

This script exports:
REGISTRY - the registry the image will be stored at.
REPOSITORY - image repository.

This script expects:
IMAGE_NAME - container disk image name.
TAG - container disk image tag.

### Example:
```bash
$ export IMAGE_NAME=example-fedora
$ export TAG=32

$ ./publish-containerdisk.sh
```

### Push the new image to local cluster registry:
```bash
$ export IMAGE_NAME=example-fedora
$ export TAG=32

# From kubevirtci / kubevirt directory
$ export REGISTRY="localhost:$(./cluster-up/cli.sh ports registry | tr -d '\r')"$

$ ./publish-containerdisk.sh
```

### Create Kubevirt VM's with the new image

In the VMI / VM yaml file, change `spec.volumes[].containerDisk.image` to the new image path.

It is possible to use an image from local cluster registry.

```bash
$ kubectl apply -f <VMI yaml file>
$ kubectl wait --for=condition=AgentConnected vmi $VMI_NAME --timeout 5m
$ virtctl console testvm1
```


## Build container-disk image

To build container-disk image form qcow2 image file, use `build-containerdisk.sh` script.

This script converts `qcow2` image to container-disk image that can be consumed by KubeVirt VM's.

What this script does:
- Creates temp directory with a copy of the source VM image.

- Build container image with kubevirt/container-disk-v1alpha layer, 
  using `Dokcerfile.template` according to:
  https://github.com/kubevirt/kubevirt/blob/master/docs/container-register-disks.md  
  The image is stored in the local registry.
  
- Exports container image as .tar file

Arguments:
- IMAGE_NAME - container-disk image name
- TAG - container-disk image tag
- VM_IMAGE_FILE - source VM image path to convert

### Example:
```bash
$ image_name='example-fedora'
$ tag='32'
$ vm_image='customized-image.qcow2'

$ ./build-containerdisk.sh $image_name $tag $vm_image
...
Successfully tagged example-fedora:32
...
Container image saved as tar file at: example-fedora_build/example-fedora-32.tar
...

# Generated image .tar file
$ ls example-fedora_build/example-fedora-32.tar
example-fedora_build/example-fedora-32.tar
```
