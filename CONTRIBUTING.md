# Contributing to kubevirt/kubevirtci

Welcome! As stated in the [README](README.md) this repository contains code for the virtualized clusters used in testing KubeVirt.

See [the KubeVirt contribution guide](https://github.com/kubevirt/kubevirt/blob/main/CONTRIBUTING.md) for general information about how to contribute.

## Development guides

- **Cluster operations** (starting, stopping, kubectl, SSH): see [K8S.md](K8S.md)
- **Local provisioning and testing**: see [KUBEVIRTCI_LOCAL_TESTING.md](KUBEVIRTCI_LOCAL_TESTING.md)
- **Phased provisioning**: see [the local testing guide](KUBEVIRTCI_LOCAL_TESTING.md#kubevirtci-phased-provision-mode)
- **Provision Manager change detection**: see [PROVISION_MANAGER.md](cluster-provision/gocli/cmd/PROVISION_MANAGER.md)
- **Pre-pulled images and pull policies**: see [the provisioning README](cluster-provision/k8s/README.md)
- **Podman support**: see [PODMAN.md](PODMAN.md)

## Repository layout

- [cluster-provision/k8s/](cluster-provision/k8s/) contains version-specific
  provisioning inputs and the shared `provision.sh` script used to build provider
  images.
- [cluster-up/cluster/](cluster-up/cluster/) contains startup provider
  configurations, including `k8s-<version>/provider.sh`. Select a startup provider
  with `KUBEVIRT_PROVIDER`, as described in [K8S.md](K8S.md).
- [cluster-provision/gocli/](cluster-provision/gocli/) contains the Go CLI used
  for cluster lifecycle operations and provisioning. Its commands include `run`,
  `provision`, `ssh`, `scp`, `ports`, `rm`, and `provision-manager`.

Provision Manager determines which providers need rebuilding. For local
provisioning, the [local testing guide](KUBEVIRTCI_LOCAL_TESTING.md)
documents `BYPASS_PMAN_CHANGE_CHECK=true` to force provisioning even when no
relevant change was detected.

## Getting started with gocli

Use the Go version and toolchain specified in
[gocli's go.mod](cluster-provision/gocli/go.mod). Container targets in the
[gocli Makefile](cluster-provision/gocli/Makefile) use `CRI_BIN` (default:
`podman`); see [PODMAN.md](PODMAN.md) for Podman setup. Single-architecture
container builds can select Docker with `CRI_BIN=docker`.

Run the following commands from the gocli directory:

```bash
cd cluster-provision/gocli
```

- `make cli`: Build the Linux CLI at `build/cli-<arch>` and create the
  `build/cli` symlink without building a container.
- `make test`: Run tests under `cmd/` and write coverage to `cover.out`.
- `make test-gocli`: Run all gocli package tests and write coverage to `cover.out`.
- `make container-run`: Build the CLI and container image, then run the image
  without a subcommand to display CLI usage.
- `make all`: Run all gocli package tests, then build and run the container image.
- `make vendor`: Run `go mod tidy` and `go mod vendor`; review resulting module
  and vendor changes before submitting them.

Use `build/cli --help` or `build/cli <command> --help` for command options.

Using local gocli images during development, and in order to test before publishing:

```bash
make container-run
export KUBEVIRTCI_GOCLI_CONTAINER=quay.io/kubevirtci/gocli:latest
```

To publish a multi-architecture image, select a tag explicitly:

```bash
make push KUBEVIRTCI_TAG=my-gocli-tag
```

`make push` builds the amd64 and s390x binaries, creates a multi-architecture
image manifest, and publishes it as `GOCLI_IMAGE:KUBEVIRTCI_TAG`.
`GOCLI_IMAGE` defaults to `KUBEVIRTCI_IMAGE_REPO/gocli`, with
`KUBEVIRTCI_IMAGE_REPO=quay.io/kubevirtci`. This target uses Podman manifest
commands and requires registry push access.

To test the published image with cluster-up, select it explicitly:

```bash
export KUBEVIRTCI_GOCLI_CONTAINER=quay.io/kubevirtci/gocli:my-gocli-tag
```

## Provider validation

After provisioning a provider, validate it with these scripts:

```bash
# Validate cluster startup
./cluster-provision/k8s/check-cluster-up.sh <provider>

# Check image pull policies
./cluster-provision/k8s/validate-pod-pull-policies.sh
./cluster-provision/k8s/validate-manifest-pull-policies.sh

# Test cluster connectivity
cluster-up/check.sh
```
