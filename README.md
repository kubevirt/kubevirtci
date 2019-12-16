# Getting Started with a multi-node Kubernetes Provider

Download this repo
```
git clone https://github.com/kubevirt/kubevirtci.git
cd kubevirtci
```

Start multi node k8s cluster with 2 nics
```
export KUBEVIRT_PROVIDER=k8s-1.13.3 KUBEVIRT_NUM_NODES=2 KUBEVIRT_NUM_SECONDARY_NICS=1
make cluster-up
```

Stop k8s cluster
```
make cluster-down
```

Use provider's kubectl client with kubectl.sh wrapper script
```
cluster-up/kubectl.sh get nodes
cluster-up/kubectl.sh get pods --all-namespaces
```

Use your own kubectl client by defining the KUBECONFIG environment variable
```
export KUBECONFIG=$(cluster-up/kubeconfig.sh)

kubectl get nodes
kubectl apply -f <some file>
```

SSH into a node
```
cluster-up/ssh.sh node01
```

Attach to node console with socat and pty (non okd4 providers)
```
# Get node01 container id
node01_id=$(docker ps |grep node01 |awk '{print $1}')

# Install socat
docker exec $node01_id yum install -y socat

# Attach to node01 console
docker exec -it $node01_id socat - /dev/pts/0
```

# Getting Started with multi-node OKD Provider

Download this repo
```
git clone https://github.com/kubevirt/kubevirtci.git
cd kubevirtci
```

Start okd cluster (pre-configured with a master and worker node)
```
export KUBEVIRT_PROVIDER=okd-4.1
make cluster-up
```

Stop okd cluster
```
make cluster-down
```

Use provider's OC client with oc.sh wrapper script
```
cluster-up/oc.sh get nodes
cluster-up/oc.sh get pods --all-namespaces
```

Use your own OC client by defining the KUBECONFIG environment variable
```
export KUBECONFIG=$(cluster-up/kubeconfig.sh)

oc get nodes
oc apply -f <some file>
```

SSH into master
```
cluster-up/ssh.sh master-0
```

SSH into worker
```
cluster-up/ssh.sh worker-0
```

Connect to the container (with KUBECONFIG exported)
```
make connect
```

Accessing OKD UI
```
TODO - in the process of working out the details here.
```

# Getting started with OCP provider scale-up

Scale up cluster with a new worker, the CUSTOM_IMAGE is in order to have a pre-provisioned qcow,
with ocp baseline images, and openshift packages.
The file will be taken from \$BASE_URL/\$CUSTOM_IMAGE.
ocp_43.repo should contain the repositories needed in order to install
openshift-ansible, openshift-clients, and openshift packages that are installed during the ansible playbook.
see please 
https://docs.openshift.com/container-platform/4.2/machine_management/adding-rhel-compute.html
for more details. and for alternative way with subscription manager.
```
export REPO_FILE=ocp_43.repo
export BASE_URL=<URL where CUSTOM_IMAGE is located>
export CUSTOM_IMAGE=custom-worker-ocp.img
make scale-up
```

How to provide a secret: \
Get your secret from https://cloud.redhat.com/openshift/install/metal/user-provisioned \
Add to it "registry.svc.ci.openshift.org" auth, save to a file and use export like this one. \
In case scale-up uses a non pre provisioned qcow image (or on provision mode), \
exporting secret must be done before cluster-up.
``` 
export INSTALLER_PULL_SECRET=$(realpath pull-secret)
```

Provision of updated qcow image: \
Every time a new provider is created, we need to update the qcow as well. \
When scale-up provision mode finishes, (it will inject known failure in ansible in order to stop it), \
copy /tmp/custom.img from the container and upload it to your server. \
update EXPECTED_MD5 with the md5sum of that image.
```
export PROVISION_MODE=1
export REPO_FILE=ocp_43.repo
export BASE_URL=<URL where CUSTOM_IMAGE is located>
export INSTALLER_PULL_SECRET=$(realpath pull-secret)
export KUBEVIRT_PROVIDER=ocp-4.3
make cluster-up
make scale-up
```

