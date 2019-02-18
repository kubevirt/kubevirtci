# KubeVirt organization images

This folder containes different base images for the projects under KubeVirt organization.

* `kubevirt-testing` - base image, that contains all needed RPM's packages for tests under the KubeVirt repository

## Versions to use

* `kubevirt-testing` - `sha256:2e43d16abaaea55672b125515e89ae69d8c6424fc2c110273aaf7db047f0b82f`

## How to modify images

1. Edit the image `Dockerfile`
2. Run `build.sh` script, it will build locally the image
3. Run `publish.sh` script, it will push the image to DockerHub under the `kubevirtci` organization
4. Update this file with updated image hash

**NOTE**: you should have write permissions for `kubevirtci` organization to push the image