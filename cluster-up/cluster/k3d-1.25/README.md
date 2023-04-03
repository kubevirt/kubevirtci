# K8s 1.25.x in a K3d cluster

Provides a pre-deployed containerized k8s cluster with version 1.25.x that runs
using [K3d](https://github.com/k3d-io/k3d)
The cluster is completely ephemeral and is recreated on every cluster restart. The KubeVirt containers are built on the
local machine and are then pushed to a registry which is exposed at
`127.0.0.1:5000`.

## Bringing the cluster up

```bash
export KUBEVIRT_PROVIDER=k3d-1.25
export KUBECONFIG=$(realpath _ci-configs/k3d-1.25/.kubeconfig)
make cluster-up
```
```
$ kubectl get nodes
NAME                 STATUS   ROLES                  AGE   VERSION
k3d-k3d-server-0   Ready    control-plane,master   67m   v1.25.6+k3s1
k3d-k3d-agent-0    Ready    worker                 67m   v1.25.6+k3s1
k3d-k3d-agent-1    Ready    worker                 67m   v1.25.6+k3s1
```

### Conneting to a node
```bash
export KUBEVIRT_PROVIDER=k3d-1.25
./cluster-up/ssh.sh <node_name> /bin/sh
```

## Bringing the cluster down

```bash
export KUBEVIRT_PROVIDER=k3d-1.25
make cluster-down
```

This destroys the whole cluster.

Note: killing the containers / cluster without gracefully moving the nics to the root ns before it,
might result in unreachable nics for few minutes.

## Using podman
Podman v4 is required.

Run:
```bash
systemctl enable --now podman.socket
ln -s /run/podman/podman.sock /var/run/docker.sock
```
The rest is as usual.
For more info see https://k3d.io/v5.4.1/usage/advanced/podman.

## Updating the provider

### Bumping K3D
Update `K3D_TAG` (see `cluster-up/cluster/k3d/common.sh` for more info)

### Bumping CNI
Update `CNI_VERSION` (see `cluster-up/cluster/k3d/common.sh` for more info)

### Bumping Multus
Download the newer manifest `https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/deployments/multus-daemonset-crio.yml`
replace this file `cluster-up/cluster/$KUBEVIRT_PROVIDER/sriov-components/manifests/multus/multus.yaml`
and update the kustomization file `cluster-up/cluster/$KUBEVIRT_PROVIDER/sriov-components/manifests/multus/kustomization.yaml`
according needs.

### Bumping calico
1. Fetch new calico yaml (https://docs.tigera.io/calico/3.25/getting-started/kubernetes/k3s/quickstart)
   Enable `allow_ip_forwarding` (See https://k3d.io/v5.4.7/usage/advanced/calico)
   Or use the one that is suggested here https://k3d.io/v5.4.7/usage/advanced/calico whenever it is updated.
2. Prefix the images in the yaml with `quay.io/` unless they have it already.
3. Update `cluster-up/cluster/k3d/manifests/calico.yaml` (see `CALICO` at `cluster-up/cluster/k3d/common.sh` for more info)

Note: Make sure to follow the latest verions on the links above.
