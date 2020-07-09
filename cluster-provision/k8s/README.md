Updating image list for pre pulling
-----------------------------------

`fetch-images.sh` can be called to extract the image identifiers from the manifests and the shell scripts in the provision dir that exists per k8s version.

However, it does **not** retrieve transitive dependencies. Therefore it can occur that during cluster up check there are images found that do not appear in the list. You can then just manually add them to the file `pre-pull-images` and therefore achieve that transitive images are then also pre pulled during provisioning.

Example:

initial pre-pull-image content:
```
calico/cni:v3.12.0
calico/kube-controllers:v3.12.0
calico/node:v3.12.0
calico/pod2daemon-flexvol:v3.12.0
fluent/fluentd-kubernetes-daemonset:v1.2-debian-syslog
fluent/fluentd:v1.2-debian
kubevirt/cdi-operator:v1.18.2
quay.io/cephcsi/rbdplugin:v1.0.0
quay.io/external_storage/local-volume-provisioner:v2.3.2
quay.io/k8scsi/csi-attacher:v1.0.1
quay.io/k8scsi/csi-node-driver-registrar:v1.0.2
quay.io/k8scsi/csi-provisioner:v1.0.1
quay.io/k8scsi/csi-snapshotter:v1.0.1
quay.io/kubevirt/cluster-network-addons-operator:0.35.0
```

During cluster-up check error:
```bash
...
Images found in cluster that are not in list!
kubevirt/cdi-operator:v1.18.2
nfvpe/multus:v3.4.1
quay.io/kubevirt/bridge-marker:0.2.0
quay.io/kubevirt/cni-default-plugins:v0.8.1
quay.io/kubevirt/kubemacpool:v0.8.3
quay.io/kubevirt/macvtap-cni:v0.2.0
quay.io/kubevirt/ovs-cni-marker:v0.11.0
quay.io/kubevirt/ovs-cni-plugin:v0.11.0
quay.io/nmstate/kubernetes-nmstate-handler:v0.17.0
...
```

After adding the above images to the list and normalizing it like so (in folder 1.18):
```bash
> mv pre-pull-images pre-pull-images.new
> cat <<EOF >> pre-pull-images.new
kubevirt/cdi-operator:v1.18.2                     
nfvpe/multus:v3.4.1                               
quay.io/kubevirt/bridge-marker:0.2.0              
quay.io/kubevirt/cni-default-plugins:v0.8.1       
quay.io/kubevirt/kubemacpool:v0.8.3               
quay.io/kubevirt/macvtap-cni:v0.2.0               
quay.io/kubevirt/ovs-cni-marker:v0.11.0           
quay.io/kubevirt/ovs-cni-plugin:v0.11.0           
quay.io/nmstate/kubernetes-nmstate-handler:v0.17.0
EOF
> bash ../fetch-images.sh 1.18 1.18/pre-pull-images.new > pre-pull-images
```

and reprovisioning the error should go away.