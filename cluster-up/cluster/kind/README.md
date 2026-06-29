# K8S in a Kind cluster

This folder serves as base to spin a k8s cluster up using [kind](https://github.com/kubernetes-sigs/kind) The cluster is completely ephemeral and is recreated on every cluster restart. 
The KubeVirt containers are built on the local machine and are then pushed to a registry which is exposed at
`localhost:5000`.

A kind cluster must specify:
* KIND_NODE_IMAGE referring the kind node image as one among those listed [here](https://hub.docker.com/r/kindest/node/tags) (please be aware that there might be compatibility issues between the kind executable and the node version)
* CLUSTER_NAME representing the cluster name 

The provider is supposed to copy a valid `kind.yaml` file under `${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/kind.yaml`

Check [kind-k8s-1.19](../kind-k8s-1.19) or [kind-1.22-sriov](kind-1.22-sriov) as examples on how to implement a kind cluster provider.

## Hotplug `misc.capacity`

Use `hotplug-misc-capacity.sh` to inject a readable
`/sys/fs/cgroup/misc.capacity` file into a running kind node for experiments.

```bash
env KUBEVIRT_PROVIDER=kind-1.34 CLUSTER_NAME=kind-1.34 ./cluster-up/up.sh
./cluster-up/cluster/kind/hotplug-misc-capacity.sh --cluster-name kind-1.34

podman exec kind-1.34-control-plane sh -c 'cat /sys/fs/cgroup/misc.capacity'
env KUBEVIRT_PROVIDER=kind-1.34 CLUSTER_NAME=kind-1.34 ./cluster-up/kubectl.sh get nodes

# If you want custom mocked content instead of the default:

./cluster-up/cluster/kind/hotplug-misc-capacity.sh \
  --cluster-name kind-1.34 \
  --content $'sev_es 100\n'

# If you want to provide a file explicitly:

printf 'sev_es 100\n' > /tmp/my-misc.capacity
./cluster-up/cluster/kind/hotplug-misc-capacity.sh \
  --cluster-name kind-1.34 \
  --source-file /tmp/my-misc.capacity
