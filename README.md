# Getting Started with a multi-node Kubernetes Provider

Download this repo
```
git clone https://github.com/kubevirt/kubevirtci.git
cd kubevirtci
```

Start multi node k8s cluster with 2 nics
```
export KUBEVIRT_PROVIDER=k8s-1.17.0 KUBEVIRT_NUM_NODES=2 KUBEVIRT_NUM_SECONDARY_NICS=1
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
# export OKD_CONSOLE_PORT=443  # Uncomment to access OKD console
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

In order to check newly created provider run,
this will point to the local created provider upon cluster-up
```
export KUBEVIRTCI_PROVISION_CHECK=1
```

# OKD Console
To access the OKD UI from the host running `docker`, remember to export `OKD_CONSOLE_PORT=443` before `make cluster-up`.
You should find out the IP address of the OKD docker container
```
clusterip=$(docker inspect $(docker ps | grep "kubevirtci/$KUBEVIRT_PROVIDER" | awk '{print $1}') | jq -r '.[0].NetworkSettings.IPAddress' )
```
and make it known in `/etc/hosts` via
```
cat << EOF >> /etc/hosts
$clusterip console-openshift-console.apps.test-1.tt.testing
$clusterip oauth-openshift.apps.test-1.tt.testing
EOF
```
Now you can browse to https://console-openshift-console.apps.test-1.tt.testing
and log in by picking the `htpasswd_provider` option. The credentials are `admin/admin`.


To access the OKD UI from a remote client, forward incoming port 433 into the OKD cluster
on the host running kubevirtci:
```
$ nic=em1  # the interface facing your remote client
$ sudo iptables -t nat -A PREROUTING -p tcp -i $nic --dport 443 -j DNAT --to-destination $clusterip
```
On your remote client host, point the cluster fqdn to the host running kubevirtci
```
kubevirtci_ip=a.b.c.d  # put here the ip address of the host running kubevirtci
cat << EOF >> /etc/hosts
$kubevirtci_ip console-openshift-console.apps.test-1.tt.testing
$kubevirtci_ip oauth-openshift.apps.test-1.tt.testing
EOF
```

# Getting started with gocli
Prerequisites:
python
Bazel

Install Bazel according https://docs.bazel.build/versions/master/install.html
Change dir to gocli folder:
```
cd cluster-provision/gocli
```

Using local gocli images durning development, and in order to test before publishing:
```
make container-run
export KUBEVIRTCI_GOCLI_CONTAINER=bazel:gocli
```

Publishing (after make container-run / make all)
```
make push 
```

After published, update cluster-up/cluster/images.sh with the gocli hash, that was created by the push command.

