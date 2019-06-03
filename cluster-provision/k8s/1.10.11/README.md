# Example Cluster

## Provisioning the cluster

Running `cli` directly:

```bash
cli provision --scripts /scripts --base kubevirtci/centos:1804_02 --tag kubevirtci/k8s-1.10.11
```

Running `cli` from within docker:

```bash
docker run --privileged --rm -v ${PWD}/scripts/:/scripts/ -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/cli provision --scripts /scripts --base kubevirtci/centos:1804_02 --tag kubevirtci/k8s-1.10.11
```

## Run the cluster

The cluster is self contained.

Running `cli` directly:

```bash
cli run --nodes 2 --base kubevirtci/k8s-1.10.11
```

Running `cli` from within docker:

```bash
docker run --privileged --rm -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/cli:latest run --nodes 2 --base kubevirtci/k8s-1.10.11
```

`--background` can be added to the `run` subcommand to exit the script after
the initial provisioning of all vms is done. If an error occures during the
initialization, the whole cluster is toren down.

## Expsosing host data via NFS

In order to share huge files (e.g. images which are only present on CI),
sharing via NFS is possible. If a directory is added via `--nfs-data` when
invoking the `run` sub-command, an additional nfs server is started and the data
can be accessed from within the VMs. The DNS name of the nfs server is `nfs`
inside the the vms.

```bash
docker run --privileged --rm -v /var/run/docker.sock:/var/run/docker.sock -v /nfs/data:/nfs/data kubevirtci/cli:latest run --nfs-data /nfs/data --nodes 2 --base kubevirtci/k8s-1.10.11
```

Within the vm it can be mounted via

```bash
sudo mount -t nfs4 nfs:/ /mnt/nfs
```

It is also very conveninent to share this data via PVs with pods this way.

## SSH into a running machine

Running `cli` directly:

```bash
cli ssh node01
```

Running `cli` from within docker:

```bash
docker run --privileged --rm -it -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/cli:latest ssh node01 
```
## Stopping the cluster

Running `cli` directly:

```bash
cli rm
```

Running `cli` from within docker:

```bash
docker run --privileged -it -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/cli:latest rm 
```


## Parallel execution

By default all the created containers will have a `kubevirt-` prefix. This way,
`cli` can detect containers which belong to it. In order to allow running
multiple clusters in parallel, a different container prefix needs to be chosen.
Every command from `cli` can be executed with `--prefix` flag to swich the
cluster and allow parallel executions.
