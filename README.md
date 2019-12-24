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

In order to check newly created provider run,
this will point to the local created provider upon cluster-up
```
export KUBEVIRTCI_PROVISION_CHECK=1
```
