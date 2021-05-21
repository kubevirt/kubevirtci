# [kubevirtci](README.md): Testing kubevirtci with locally provisioned cluster

After making changes to a kubevirtci provider, you should test it locally before publishing it.

Let's go on the required steps, starting by provisioning, all the way to `make cluster-up`.
`$KUBEVIRTCI_DIR` is assumed to be your kubevirtci path.

Steps:

### kubevirtci: provision cluster locally

```bash
# switch to kubevirtci directory
cd $KUBEVIRTCI_DIR
```

```bash
# Build a provider. This includes starting it with cluster-up for verification and shutting it down for cleanup.
(cd cluster-provision/k8s/1.19; ../provision.sh)
```

### prepare for using the new provisioned cluster

```bash
# set local provision test flag (mandatory)
export KUBEVIRTCI_PROVISION_CHECK=1
```

Note:
In case you updated gocli and need to test it locally as well, export additionally:
```bash
export KUBEVIRTCI_GOCLI_CONTAINER=quay.io/kubevirtci/gocli:latest
```

### start cluster

```bash
export KUBEVIRT_PROVIDER=k8s-1.19
export KUBECONFIG=$(./cluster-up/kubeconfig.sh)
export KUBEVIRT_NUM_NODES=2

# spin up cluster
make cluster-up
```

#### start cluster with prometheus, alertmanager and grafana
Prometheus operator is only supported for providers >= 1.21. To enable prometheus, please also export the following variables before running `make cluster-up`:
```bash
export KUBEVIRT_PROVIDER=k8s-1.21
export KUBEVIRT_DEPLOY_PROMETHEUS=true
export KUBEVIRT_DEPLOY_PROMETHEUS_ALERTMANAGER=false
export KUBEVIRT_DEPLOY_GRAFANA=true
```


## kubevirt: testing kubevirt locally with a freshly provisioned cluster

After making changes to a kubevirtci provider, it's recommended to test it locally including kubevirt e2e tests before publishing it.

With the changes in place you can execute locally [`make functest`](https://github.com/kubevirt/kubevirt/blob/master/docs/getting-started.md#testing) against a cluster with kubevirt that was provisioned using `kubevirtci`.

`$KUBEVIRT_DIR` is assumed to be your kubevirt path.

Steps:

### sync cluster-up folder

```bash
# sync _ci-configs folder (mandatory, since it has data about the current running cluster).
rsync -av $KUBEVIRTCI_DIR/_ci-configs/ $KUBEVIRT_DIR/_ci-configs
# sync cluster-up folder if it has changed.
rsync -av $KUBEVIRTCI_DIR/cluster-up/ $KUBEVIRT_DIR/cluster-up
```

```bash
# switch directory to kubevirt folder
cd $KUBEVIRT_DIR
```

### run kubevirt tests

```bash
# deploy latest kubevirt changes to cluster
make cluster-sync

# start tests, either
make functest

# or use ginkgo focus
FUNC_TEST_ARGS='-focus=vmi_cloudinit_test -regexScansFilePath' make functest
```