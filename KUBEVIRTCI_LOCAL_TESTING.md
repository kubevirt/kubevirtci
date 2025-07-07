# [kubevirtci](README.md): Testing kubevirtci with locally provisioned cluster

After making changes to a kubevirtci provider, you should test it locally before publishing it.

Let's go on the required steps, starting by provisioning, all the way to `make cluster-up`.
`$KUBEVIRTCI_DIR` is assumed to be your kubevirtci path.

Steps:

### kubevirtci: provision cluster locally

Install the following CLI tools before running the script:
* `docker` or `podman`
* `go`
* `kubectl`

```bash
# switch to kubevirtci directory
cd $KUBEVIRTCI_DIR
```

```bash
# Build a provider. This includes starting it with cluster-up for verification and shutting it down for cleanup.
(cd cluster-provision/k8s/1.33; ../provision.sh)
```

Note: 
If you see "INFO: skipping provision of x.yz because according provision-manager it hadn't changed"
please use `export BYPASS_PMAN_CHANGE_CHECK=true` to bypass provision-manager changes check and force provision.

### prepare for using the new provisioned cluster

```bash
# set local provision test flag (mandatory)
export KUBEVIRTCI_PROVISION_CHECK=1
```
This ensures to set container-registry to quay.io and container-suffix to :latest
If `KUBEVIRTCI_PROVISION_CHECK` is not used,
you can set `KUBEVIRTCI_CONTAINER_REGISTRY` (default: `quay.io`), `KUBEVIRTCI_CONTAINER_ORG` (default: `kubevirtci`) and `KUBEVIRTCI_CONTAINER_SUFFIX` (default: according gocli tag),
in order to use a custom image.

Note:
In case you updated gocli and need to test it locally as well, export additionally:
```bash
export KUBEVIRTCI_GOCLI_CONTAINER=quay.io/kubevirtci/gocli:latest
```

### start cluster

```bash
export KUBEVIRT_PROVIDER=k8s-1.33
export KUBECONFIG=$(./cluster-up/kubeconfig.sh)
export KUBEVIRT_NUM_NODES=2

# spin up cluster
make cluster-up
```

#### start cluster with prometheus, alertmanager and grafana
To enable prometheus, please also export the following variables before running `make cluster-up`:
```bash
export KUBEVIRT_PROVIDER=k8s-1.33
export KUBEVIRT_DEPLOY_PROMETHEUS=true
export KUBEVIRT_DEPLOY_PROMETHEUS_ALERTMANAGER=true
export KUBEVIRT_DEPLOY_GRAFANA=true
```

#### start cluster with swap enabled
To enable swap, please also export the following variables before running `make cluster-up`:
```bash
# to tune swap:
# KUBEVIRT_SWAP_SIZE_IN_GB - Change the swap file size 
# the default size is 2GB
# KUBEVIRT_KSM_PAGES_TO_SCAN - The swappiness parameter determines how aggressively 
# the kernel will swap out memory pages.
# values are between 0-100 if the value is higher than the kernel will more aggressive
# the default value is 30
# KUBEVIRT_UNLIMITEDSWAP - Kubernetes workloads can use as much swap memory as they 
# request, up to the system limit (without consideration to the pod's memory limit)
export KUBEVIRT_SWAP_ON=true
```

#### start cluster with ksm enabled
To enable KSM (Kernel Samepage Merging), please also export the following variables before running `make cluster-up`:
```bash
# to tune ksm:
# KUBEVIRT_KSM_SLEEP_BETWEEN_SCANS_MS - This parameter controls
# how long KSM should sleep in millisecond between scans.
# the Default value is 20
# KUBEVIRT_KSM_PAGES_TO_SCAN - This parameter controls how many
# pages KSM should scan in each pass.
# The default value for pages_to_scan is 100, which means each 
# scan run only inspects about half a megabyte of RAM.
export KUBEVIRT_KSM_ON=true
```
For more details see pages_to_scan and sleep_millisecs in: https://www.kernel.org/doc/Documentation/vm/ksm.txt

## kubevirt: testing kubevirt locally with a freshly provisioned cluster

After making changes to a kubevirtci provider, it's recommended to test it locally including kubevirt e2e tests before publishing it.

With the changes in place you can execute locally [`make functest`](https://github.com/kubevirt/kubevirt/blob/main/docs/getting-started.md#testing) against a cluster with kubevirt that was provisioned using `kubevirtci`.

`$KUBEVIRT_DIR` is assumed to be your kubevirt path.

Steps:

### sync cluster-up folder

```bash
# sync _ci-configs folder (mandatory, since it has data about the current running cluster).
rsync -av $KUBEVIRTCI_DIR/_ci-configs/ $KUBEVIRT_DIR/kubevirtci/_ci-configs
# sync cluster-up folder if it has changed.
rsync -av $KUBEVIRTCI_DIR/cluster-up/ $KUBEVIRT_DIR/kubevirtci/cluster-up
```

```bash
# switch directory to kubevirt folder
cd $KUBEVIRT_DIR
```

### kubevirtci phased provision mode

Kubevirtci supports phased provision mode in order to save time while developing.  
There are two phases:  
`linux` phase which updates the VM kernel, install required packages such as cri-o,
pre pull images and configure the OS.  
`k8s` phase which configures the network and creates the cluster.

The default mode is to do both phases in the same flow and then to run cluster check.
Sometimes we need to repeat only the `k8s` phase, and then to test it locally once we stabilize it.

For that we have phased mode.
Usage: export the required mode, i.e `export PHASES=linux` or `export PHASES=k8s`
and then run the provision. the full flow will be:

`export PHASES=linux; (cd cluster-provision/k8s/1.33; ../provision.sh)`  
`export PHASES=k8s; (cd cluster-provision/k8s/1.33; ../provision.sh)`  
Run the `k8s` step as much as needed. It reuses the intermediate image that was created
by the `linux` phase.
Note :
1. By default when you run `k8s` phase alone, it uses centos9 image specified in cluster-provision/k8s/base-image, not the one built locally in the `linux` phase. So, to make `k8s` phase use the locally built centos9 image, update cluster-provision/k8s/base-image with the locally built image name and tag (default: quay.io/kubevirtci/centos9:latest)
2. Also note if you run both `linux,k8s` phases, then it doesn't save the intermediate container image generated post linux image. So, for the centos9 image required for k8s stage, you've to run the linux phase alone.

Once you are done, either check the cluster manually, or use:  
`export PHASES=k8s; export CHECK_CLUSTER=true; (cd cluster-provision/k8s/1.33; ../provision.sh)`

### provision without pre-pulling images

In order to develop faster, you can skip pre-pulling the optional images,
such as CDI, CNAO, Prometheus, Istio and so on.  
Run `export SLIM=true` before provisioning, to create such provider.

### run kubevirt tests

```bash
# deploy latest kubevirt changes to cluster
make cluster-sync

# start tests, either
make functest

# or use ginkgo focus
FUNC_TEST_ARGS='-focus=vmi_cloudinit_test -regexScansFilePath' make functest
```
