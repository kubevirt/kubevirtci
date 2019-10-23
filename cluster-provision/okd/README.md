# How to create new OKD release

Possible the situatution when the specific OKD componenet need have some bug fix, that does not exist under the release image, in this case you can build new component image with the fix and create a new release image that will use this component image.

1. You will need to download [`oc`](https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/) binary and extract it.

2. After you can check all images via `oc adm release info quay.io/openshift-release-dev/ocp-release:4.1.18`

3. And create a new image that will use a new component image `oc adm release new --to-image docker.io/kubevirtci/ocp-release:4.1.18 --from-release quay.io/openshift-release-dev/ocp-release:4.1.18 libvirt-machine-controllers=docker.io/kubevirtci/origin-libvirt-machine-controllers@sha256:090d4035c6558cdc956d5fed70b0646998c9c4058ed1791370d76d8553130244`

***
Note: Be sure that you have `pull` permissions for the image repository, in the case of above example, you should have permissions for `docker.io/kubevirtci` and you will need read permissions for `quay.io/openshift-release-dev` repository.
***
