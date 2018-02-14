# Example Cluster

## Provisioning the cluster

Running `cli` directly:

```bash
cli provision --scripts /scripts --base rmohr/centos:1608_01 --tag rmohr/kubeadm-1.9.3
```

Running `cli` from withing docker:

```bash
docker run --privileged --rm -v ${PWD}/scripts/:/scripts/ -v /var/run/docker.sock:/var/run/docker.sock rmohr/cli provision --scripts /scripts --base rmohr/centos:1608_01 --tag rmohr/kubeadm-1.9.3
```

## Run the cluster

The cluster is self contained.

Running `cli` directly:

```bash
cli run --nodes 2 --base rmohr/kubeadm-1.9.3
```

Running `cli` from withing docker:

```bash
docker run --privileged --rm -v /var/run/docker.sock:/var/run/docker.sock rmohr/cli:latest run --nodes 2 --base rmohr/kubeadm-1.9.3
```
## SSH into a running machine

Running `cli` directly:

```bash
cli ssh node01
```

Running `cli` from withing docker:

```bash
docker run --privileged --rm -it -v /var/run/docker.sock:/var/run/docker.sock rmohr/cli:latest ssh node01 
```
## Stopping the cluster

Running `cli` directly:

```bash
cli rm
```

Running `cli` from withing docker:

```bash
docker run --privileged -it -v /var/run/docker.sock:/var/run/docker.sock rmohr/cli:latest rm 
```

## Parallel execution

By default all the created containers will have a `kubevirt-` prefix. This way,
`cli` can detect containers which belong to it. In order to allow running
multiple clusters in parallel, a different container prefix needs to be chosen.
Every command from `cli` can be executed with `--prefix` flag to swich the
cluster and allow parallel executions.
