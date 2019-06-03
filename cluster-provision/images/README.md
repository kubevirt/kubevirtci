# KubeVirt organization images

This folder containes different base images for the projects under KubeVirt organization.

* `kubevirt-testing` - base image, that contains all needed RPM's packages for tests under the KubeVirt repository

## Versions to use

* `kubevirt-testing` - `sha256:eb86f7388217bb18611c8c4e6169af3463c2a18f420314eb4d742b3d3669b16f`

## How to modify images

1. Edit the image `Dockerfile`
2. Run `build.sh` script, it will build locally the image
3. Run `publish.sh` script, it will push the image to DockerHub under the `kubevirtci` organization
4. Update this file with updated image hash

**NOTE**: you should have write permissions for `kubevirtci` organization to push the image
