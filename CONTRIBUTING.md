# Contributing to kubevirt/kubevirtci

Welcome! As stated in the [README](README.md) this repository contains code for the virtualized clusters used in testing KubeVirt.

See [the KubeVirt contribution guide](https://github.com/kubevirt/kubevirt/blob/main/CONTRIBUTING.md) for general information about how to contribute.

## Development guides

- **Cluster operations** (starting, stopping, kubectl, SSH): see [K8S.md](K8S.md)
- **Local provisioning and testing**: see [KUBEVIRTCI_LOCAL_TESTING.md](KUBEVIRTCI_LOCAL_TESTING.md)
- **Podman support**: see [PODMAN.md](PODMAN.md)

## Getting started with gocli

```
cd cluster-provision/gocli
```

Using local gocli images during development, and in order to test before publishing:
```
make container-run
export KUBEVIRTCI_GOCLI_CONTAINER=quay.io/kubevirtci/gocli:latest
```

Publishing (after make container-run / make all)
```
make push
```

After published, update cluster-up/cluster/images.sh with the gocli hash, that was created by the push command.
Or simply use:
```
make bump provider=gocli hash=<NEW_HASH>
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

