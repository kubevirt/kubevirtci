# k8s clusters in QEMU in docker

* `base` contains the base image with some scripts, qemu and dnsmasq
* `centos7` adds a vagrant centos7 box to the image
* `cli` contains a tool for provisioning, running and managing the containerized clusters
* `k8s-1.10.11` k8s-1.10.11 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-1.11.0` k8s-1.11.0 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-1.13.3` k8s-1.13.3 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-multus-1.13.3` k8s-1.13.3 cluster based on the centos7 image and uses multus CNI, provisioned with kubeadm
* `k8s-genie-1.11.1` k8s-1.11.1 cluster based on the centos7 image and uses genie CNI, provisioned with kubeadm
* `os-3.11.0` os-3.11.0 cluster based on the centos7 image, provisioned with openshift-ansible
* `os-3.11-crio` os-3.11 cluster with CRI-O support based on the centos7 image, provisioned
* `os-3.11-multus` os-3.11 cluster with multus cni support based on the centos7 image, provisioned with openshift-ansible

## Versions to use

* `kubevirtci/cli`: `sha256:1dd015dea4f12e6dcb8e31be3eeb677fed96f290ef4a4892a33c43d666053536`
* `kubevirtci/gocli`: `sha256:868fd9f2b6e5ff6473ab92cdd7d3a762ec8a8c66dac29f5db841879d40038f2a`
* `kubevirtci/base`: `sha256:1f9f2e12571cf96a89a3e24b15de0b2857db159b4c7d3657c81541845040c1bd`
* `kubevirtci/centos:1905_01`: `sha256:728232c8b5b79bb69e1386a6b47acd72876faa386202b535e227bc605adffe33`
* `kubevirtci/os-3.11.0-multus`: `sha256:adabdd119fe04d1808c57418369ea7ec84b94770f1686ca168f079c6618b23a4`
* `kubevirtci/os-3.11.0:`: `sha256:ff54198c7eaa51ccf2aeff6d5dd1e9ca51ccfec6ee0237b0305fb3b6761f95f6`
* `kubevirtci/os-3.11.0-crio:`: `sha256:66b348692b70c8017ef4e914f08252de5c26145640c516f749d38fc78c666589`
* `kubevirtci/k8s-1.10.11:`: `sha256:f563a8ab4719e53c2372c4f41dfe55256677ec7afc442dfaebd494926005e3e5`
* `kubevirtci/k8s-1.11.0:`: `sha256:afa15dbea2b520e7c5abe5880127bbfd5a9c15207792082d2cc6409943890001`
* `kubevirtci/k8s-1.13.3:`: `sha256:40cc8c51bfcf87eb0e206ad38c3319a5a26ea479f71ee63195705f7dae59e00d`
* `kubevirtci/k8s-multus-1.13.3:`: `sha256:6e7245922c44d06374534bbce7981934a964efd075e6de44bd532d869f267811`
* `kubevirtci/k8s-genie-1.11.1:`: `sha256:b63ec1515bd88ee39dc46159c32fa4ee23c890005cbda7745a90dd9b72c25254`

# OKD clusters in the container with libvirt

* `okd-base` contains all needed packages to provision and run OKD cluster on top of the libvirt provider
* `okd-4.1.0` okd-4.1.0 cluster provisioned with OpenShift installer on top of the libvirt provider
* `okd-4.1.2` okd-4.1.2 cluster provisioned with OpenShift installer on top of the libvirt provider, this image contains custom libvirt image that includes fixes to deploy new nodes without need to apply any W/A

## Versions to use

* `kubevirtci/okd-base`: `sha256:259e776998da3a503a30fdf935b29102443b24ca4ea095c9478c37e994e242bb`
* `kubevirtci/okd-4.1.0`: `sha256:d190cee4bb30e231ceb9a7c9eb1ade10c036225e126cd0abf60e9706ebd696fd`
* `kubevirtci/okd-4.1.2`: `sha256:88fba5c00ba973c8da712d14689f1d93c40fa6a8e8efdb4da501b572adbd3d6b`

## Using gocli

`gocli` is a tiny go binary which helps managing the containerized clusters. It
ca be used from a docker images, so no need to install it. You can for instance
use a bash alias:

```bash
alias gocli="docker run --net=host --privileged --rm -it -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/gocli:latest"
gocli help
```

### How to provision OKD cluster

First you will need to create installer pull token file with the content:
```
pullSecret: <pull secret>
```

and after you should run `gocli` command:
```bash
gocli provision okd \
--prefix okd-4.1.0 \
--dir-scripts <scripts_folder>/scripts \
--dir-hacks <hacks_folder>/hacks \
--master-memory 10240 \
--installer-pull-token-file <installer_pull_token_file> \
--installer-repo-tag release-4.1 \
--installer-release-image quay.io/openshift-release-dev/ocp-release:4.1.0-rc.7 \
kubevirtci/okd-base@sha256:259e776998da3a503a30fdf935b29102443b24ca4ea095c9478c37e994e242bb
```

***
NOTE: you can get the pull secret [here](https://developers.redhat.com/auth/realms/rhd/protocol/openid-connect/auth?client_id=uhc&redirect_uri=https%3A%2F%2Fcloud.openshift.com%2Fclusters%2Finstall%23pull-secret&state=109aa48e-1779-45d6-9bdc-c156b1e699b4&response_mode=fragment&response_type=code&scope=openid&nonce=b9fe0085-f2c9-4fd7-bd17-e8629f01bbaf).
***

***
NOTE: OpenShift cluster consumes a lot of resources, you should have at least 18Gb of the memory on the machine where do you run the container.
***

### How to run OKD cluster

You should run `gocli` command:
```bash
gocli run okd --prefix okd-4.1.0 --ocp-console-port 443 --background kubevirtci/okd-4.1.0@sha256:d190cee4bb30e231ceb9a7c9eb1ade10c036225e126cd0abf60e9706ebd696fd
```

### How to connect to the OKD console

To connect the OKD console you should add once hosts to the `/etc/hosts`

```bash
127.0.0.1 console-openshift-console.apps.test-1.tt.testing
127.0.0.1 oauth-openshift.apps.test-1.tt.testing
```

and specify the `--ocp-console-port` under the `gocli` run command to `443`.

After you can connect to the https://console-openshift-console.apps.test-1.tt.testing and login via `htpasswd_provider` provider with `admin` user and password.

## Quickstart Kubernetes

### Start the cluster

Start a k8s cluster which contains of one master and two nodes:

```bash
gocli run --random-ports --nodes 3 --background kubevirtci/k8s-1.13.3
```

### Connect to the cluster

Find out the connection details of the cluster:

```bash
$ gocli ports k8s
33396
$ gocli scp /etc/kubernetes/admin.conf - > ./kubeconfig
$ kubectl --kubeconfig ./kubeconfig --insecure-skip-tls-verify --server https://localhost:33396 get pods -n kube-system
NAME                             READY     STATUS    RESTARTS   AGE
etcd-node01                      1/1       Running   0          14m
kube-apiserver-node01            1/1       Running   0          13m
kube-controller-manager-node01   1/1       Running   0          14m
kube-dns-6f4fd4bdf-mh6nb         3/3       Running   0          14m
kube-flannel-ds-4bk76            1/1       Running   0          14m
kube-flannel-ds-5zgmt            1/1       Running   1          14m
kube-flannel-ds-qbm2r            1/1       Running   1          14m
kube-proxy-gtvpb                 1/1       Running   0          14m
kube-proxy-knc6p                 1/1       Running   0          14m
kube-proxy-vx9t6                 1/1       Running   0          14m
kube-scheduler-node01            1/1       Running   0          13m
```

or to permamently edit kubeconfig:

```bash
$ gocli scp /etc/kubernetes/admin.conf - > ./kubeconfig
$ kubectl --kubeconfig=./kubeconfig config set-cluster kubernetes --server=https://127.0.0.1:$(gocli ports k8s|tr -d '\r\n')
$ kubectl --kubeconfig=./kubeconfig config set-cluster kubernetes --insecure-skip-tls-verify=true
$ kubectl --kubeconfig ./kubeconfig get pods -n kube-system
NAME                             READY     STATUS    RESTARTS   AGE
etcd-node01                      1/1       Running   0          14m
kube-apiserver-node01            1/1       Running   0          13m
kube-controller-manager-node01   1/1       Running   0          14m
kube-dns-6f4fd4bdf-mh6nb         3/3       Running   0          14m
kube-flannel-ds-4bk76            1/1       Running   0          14m
kube-flannel-ds-5zgmt            1/1       Running   1          14m
kube-flannel-ds-qbm2r            1/1       Running   1          14m
kube-proxy-gtvpb                 1/1       Running   0          14m
kube-proxy-knc6p                 1/1       Running   0          14m
kube-proxy-vx9t6                 1/1       Running   0          14m
kube-scheduler-node01            1/1       Running   0          13m
```

### Destroy the cluster

```bash
$ gocli rm
```

## Quickstart OpenShift

### Start the cluster

Start a k8s cluster which contains of one master and two nodes:

```bash
gocli run --random-ports --nodes 2 --memory 5120M --reverse --ocp-port 8443 --background kubevirtci/os-3.11.0
```

Note the extra `--reverse` flag. Normally we start the master first and nodes
register. In the case of openshift it is different. We first need to start the
nodes, so that openshift-ansible can reach out to the nodes from master.

Furter we added `--ocp-port 8443`. That is only required if you want to access
the openshift-web-console. More about that further below.

### Connect to the cluster

Find out the connection details of the cluster:

```bash
$ gocli ports k8s
8443
$ gocli scp /etc/origin/master/admin.kubeconfig - > ./kubeconfig
$ oc --kubeconfig=./kubeconfig config set-cluster node01:8443 --server=127.0.0.1:8443
$ oc --kubeconfig=./kubeconfig config set-cluster node01:8443 --insecure-skip-tls-verify=true
$ oc --kubeconfig ./kubeconfig get nodes
```

### Accessing the webconsole

Make sure that `node01` resolves to `127.0.0.1` and that you added `--ocp-port
8443` when creatin the cluster. If you did that, you can simply access the
webconsole at `https://127.0.0.1:8443`. The login credentials are
`admin:admin`.

The two preconditions are necessary to make the authentication redirects work.

### Destroy the cluster

```bash
$ gocli rm
```
