# DRA vfio-gpu for KubeVirt e2e

Provides fake `vfio-pci` devices for KubeVirt DRA e2e tests on the
virtualized `k8s-*` providers. The synthetic PCI devices come from the
`fake-iommu` and `fake-pci` kernel modules in the dra-example-driver
checkout (`demo/vfio-gpu`). This directory clones that repo and builds the
modules inside the virtualized worker nodes instead of on the host.

> Inspired by the approach explored in
> [kubevirt/kubevirt#16712](https://github.com/kubevirt/kubevirt/pull/16712)
> (synthetic kernel devices for GPU e2e testing without real hardware).

> The host still needs to be able to run kubevirtci `k8s-*` providers with KVM,
> but the fake VFIO kernel modules are loaded in the provider VMs.

> **WARNING:** loading and binding the fake VFIO kernel modules requires
> `root`. The `k8s-*` provider path loads the fake modules inside the
> provider worker nodes. For a Kind cluster, run
> `demo/vfio-gpu/setup-host-vfio-pci.sh` from a dra-example-driver checkout.

## Contents

| File / dir | Purpose |
| ---------- | ------- |
| `config_vfio_cluster.sh` | Post-create cluster setup: copy `demo/vfio-gpu` into each worker, label nodes, install the DRA driver |
| `install_dra_example_driver.sh` | Clones, builds, pushes, and installs the DRA example driver |
| `../k8s-*/config_vfio_cluster.sh` | Provider wrapper gated by `KUBEVIRT_USE_FAKE_VFIO=true` |

The kernel modules and `setup-node-vfio.sh` live in the driver's `demo/vfio-gpu` directory. `config_vfio_cluster.sh` clones that repo first and copies the directory into each worker.

## Prerequisites

Host tools: `docker` or `podman`, `kubectl`, `helm`, `git`, and a
kubevirtci `k8s-*` provider that can run nested virtualization. The modules
are built inside the worker, so the host does not need kernel headers.

Worker node packages: `make`, `gcc`, and matching kernel headers for the
running worker-node kernel. These must be present in the provider image; the
VFIO setup validates them before building the fake modules.

Memory requirements: Some fake VFIO tests require at least **8G (8192M)** of
memory allocated to the node VM. The default `KUBEVIRT_MEMORY_SIZE` (`5120M`)
is insufficient — `make cluster-up` will **fail** with a fatal error if
`KUBEVIRT_MEMORY_SIZE` is below this threshold when `KUBEVIRT_USE_FAKE_VFIO=true`.
`KUBEVIRT_MEMORY_SIZE` accepts a `m`/`M` (mebibytes) or `g`/`G` (gibibytes)
suffix, e.g. `8192M` or `8G`; a value with no suffix is treated as mebibytes.

## Cluster setup


```bash
export KUBEVIRT_PROVIDER=k8s-1.36
export KUBEVIRT_USE_FAKE_VFIO=true
export KUBEVIRT_MEMORY_SIZE=8192M
export FAKE_PCI_DEVICES=8
export FAKE_IOMMU=true

make cluster-up
```

`make cluster-up` creates the `k8s-*` provider and, when
`KUBEVIRT_USE_FAKE_VFIO=true`, runs the provider's `config_vfio_cluster.sh`
wrapper. That wrapper delegates to this directory and configures each worker
node by:

- cloning dra-example-driver and copying `demo/vfio-gpu` to `/tmp/fake-vfio`;
- building and loading `fake-iommu.ko` and `fake-pci.ko` inside the worker;
- binding fake PCI devices to `vfio-pci`;
- labeling nodes with `fake-vfio-capable=true`;
- building, pushing, and installing the DRA example driver.

CPU Manager must be enabled for the KubeVirt e2e setup. The `k8s-*` worker
bootstrap configures static CPU Manager policy for supported architectures.

To rerun only the VFIO/DRA setup against an existing cluster:

```bash
bash cluster-up/cluster/k8s-1.36/config_vfio_cluster.sh
```

## Kind cluster testing

If you want to test against your own Kind cluster, load and bind the fake VFIO
kernel modules on the host first, from a dra-example-driver checkout:

```bash
sudo -E FAKE_PCI_DEVICES=8 bash demo/vfio-gpu/setup-host-vfio-pci.sh
```

Create the Kind cluster with `/dev/vfio` bind-mounted into the node container,
then make the VFIO control device writable inside the node:

```yaml
nodes:
- role: control-plane
  extraMounts:
  - hostPath: /dev/vfio/
    containerPath: /dev/vfio/
```

```bash
docker exec <kind-node> mount -o remount,rw /sys
docker exec <kind-node> chmod 666 /dev/vfio/vfio
```

The `k8s-*` `config_vfio_cluster.sh` flow does not run for an existing Kind
cluster unless a Kind provider explicitly calls it.

## Configuration

| Variable | Default | Purpose |
| -------- | ------- | ------- |
| `KUBEVIRT_USE_FAKE_VFIO` | `false` | Enables fake VFIO setup during `make cluster-up` for supported `k8s-*` providers |
| `KUBEVIRT_MEMORY_SIZE` | `5120M` | Memory allocated to provider node VMs. Accepts a `m`/`M` or `g`/`G` suffix (e.g. `8192M`, `8G`). Set to `>= 8192M` (`8G`) for fake VFIO |
| `FAKE_PCI_DEVICES` | `8` | Number of synthetic PCI devices to create inside each worker node |
| `FAKE_IOMMU` | `true` | Load the fake IOMMU companion so fake devices can bind to `vfio-pci` |
| `DRA_DRIVER_PROFILE` | `vfio-gpu` | DRA example driver profile installed by Helm |
| `DRA_DRIVER_NAME` | `vfio-gpu.example.com` | DRA driver name exposed to KubeVirt |
| `DRA_DRIVER_IMAGE_NAME` | `dra-example-driver` | Image name used for the locally built DRA driver |
