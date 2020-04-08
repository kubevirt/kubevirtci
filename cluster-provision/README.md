# k8s clusters in QEMU in docker

* `centos7` adds a vagrant centos7 box to the image
* `cli` contains a tool for provisioning, running and managing the containerized clusters
* `k8s-1.10.11` k8s-1.10.11 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-1.11.0` k8s-1.11.0 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-1.13.3` k8s-1.13.3 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-multus-1.13.3` k8s-1.13.3 cluster based on the centos7 image and uses multus CNI, provisioned with kubeadm
* `os-3.11.0` os-3.11.0 cluster based on the centos7 image, provisioned with openshift-ansible
* `os-3.11-crio` os-3.11 cluster with CRI-O support based on the centos7 image, provisioned
* `os-3.11-multus` os-3.11 cluster with multus cni support based on the centos7 image, provisioned with openshift-ansible

## Versions to use

* `kubevirtci/cli`: `sha256:1dd015dea4f12e6dcb8e31be3eeb677fed96f290ef4a4892a33c43d666053536`
* `kubevirtci/gocli`: `sha256:220f55f6b1bcb3975d535948d335bd0e6b6297149a3eba1a4c14cad9ac80f80d`
* `kubevirtci/centos:1905_01`: `sha256:4b292b646f382d986c75a2be8ec49119a03467fe26dccc3a0886eb9e6e38c911`
* `kubevirtci/centos:2001_01`: `sha256:6f2548dcc23489d0c945aef516781ae2ea678424c3760d1dafa0a83d29411713`
* `kubevirtci/os-3.11.0-multus`: `sha256:0c8be10799490a1f86740eaa490063f51eab78b920540f0a2946abc5e0bf30fe`
* `kubevirtci/os-3.11.0:`: `sha256:ebc9048f25eb5cc720b8b4eeab7b58b5fa3648d27c9612d87bf338d5dbee46a7`
* `kubevirtci/os-3.11.0-crio:`: `sha256:71ea794ff45e06ac521e2fe867f192b98ae989755629e4830ab919ecd3905337`
* `kubevirtci/k8s-1.14.6`: `sha256:ec29c07c94fce22f37a448cb85ca1fb9215d1854f52573316752d19a1c88bcb3`
* `kubevirtci/k8s-1.15.1`: `sha256:14d7b1806f24e527167d2913deafd910ea46e69b830bf0b094dde35ba961b159`
* `kubevirtci/k8s-1.16.2`: `sha256:5bae6a5f3b996952c5ceb4ba12ac635146425909801df89d34a592f3d3502b0c`
* `kubevirtci/k8s-multus-1.13.3:`: `sha256:c0bcf0d2e992e5b4d96a7bcbf988b98b64c4f5aef2f2c4d1c291e90b85529738`

# OKD clusters in the container with libvirt

* `okd-base` contains all needed packages to provision and run OKD cluster on top of the libvirt provider
* `okd-4.1` okd-4.1 cluster provisioned with OpenShift installer on top of the libvirt provider, this image contains custom libvirt image that includes fixes to deploy new nodes without need to apply any W/A
* `okd-4.3` okd-4.3 cluster provisioned with OpenShift installer on top of the libvirt provider, this image contains custom libvirt image that includes fixes to deploy new nodes without need to apply any W/A

## Versions to use

* `kubevirtci/okd-base`: `sha256:73ede51ce464546a82b81956b7f58cf98662a4c5fded9c659b57746bc131e047`
* `kubevirtci/okd-4.1`: `sha256:e7e3a03bb144eb8c0be4dcd700592934856fb623d51a2b53871d69267ca51c86`
* `kubevirtci/okd-4.2`: `sha256:a830064ca7bf5c5c2f15df180f816534e669a9a038fef4919116d61eb33e84c5`
* `kubevirtci/okd-4.3`: `sha256:63abc3884002a615712dfac5f42785be864ea62006892bf8a086ccdbca8b3d38`

## Using gocli

`gocli` is a tiny go binary which helps managing the containerized clusters. It
can be used from a docker images, so no need to install it. You can for instance
use a bash alias:

```bash
alias gocli="docker run --net=host --privileged --rm -it -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/gocli:latest"
gocli help
```

### How to provision OKD cluster

First you will need to create installer pull token file with the content:
```
{"auths":{...}}
```

and after you should run `gocli` command:
```bash
gocli provision okd \
--prefix okd-4.1 \
--dir-scripts <scripts_folder>/scripts \
--dir-hacks <hacks_folder>/hacks \
--master-memory 10240 \
--installer-pull-secret-file <installer_pull_secret_file> \
--installer-repo-tag release-4.1 \
--installer-release-image quay.io/openshift-release-dev/ocp-release:4.1 \
kubevirtci/okd-base@sha256:73ede51ce464546a82b81956b7f58cf98662a4c5fded9c659b57746bc131e047
```

***
NOTE: you can get the pull secret [here](https://cloud.redhat.com/openshift).
***

***
NOTE: OpenShift cluster consumes a lot of resources, you should have at least 18Gb of the memory on the machine where do you run the container.
***

### How to run OKD cluster

You should run `gocli` command:
```bash
gocli run okd --prefix okd-4.1 --ocp-console-port 443 --installer-pull-secret-file <installer_pull_secret_file> --background kubevirtci/okd-4.1@sha256:e7e3a03bb144eb8c0be4dcd700592934856fb623d51a2b53871d69267ca51c86
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
