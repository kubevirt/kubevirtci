# [kubevirtci](README.md): Testing kubevirt with locally provisioned cluster

With the changes in place you also can directly execute [`make functest`](https://github.com/kubevirt/kubevirt/blob/master/docs/getting-started.md#testing) against a cluster locally with kubevirt that was provisioned using `kubevirtci`.

Steps:

### kubevirtci: provision cluster locally

```bash
cd .../kubevirtci
cd cluster-provision/k8s/1.17.0
../provision.sh      # this also calls cluster-up to check whether the cluster will really start and have the pods ready
```

### kubevirt: prepare for tests

```bash
# set local provision test flag
export KUBEVIRTCI_PROVISION_CHECK=1
# sync changes to kubevirt cluster-up, notably cluster/images.sh
rsync -av .../kubevirtci/cluster-up .../kubevirt/cluster-up                               
```                                                                                       
                                                                                          
### start cluster and test                                                                
                                                                                          
```bash                                                                                   
cd .../kubevirt                                                                           
                                                                                          
# spin up cluster                                                                         
make cluster-up                                                                           
                                                                                          
# deploy latest kubevirt changes to cluster                                               
make cluster-sync                                                                         
                                                                                          
# start tests, either                                                                     
make functest                                                                             
                                                                                          
# or use ginkgo focus                                                                     
FUNC_TEST_ARGS='-ginkgo.focus=vmi_cloudinit_test -ginkgo.regexScansFilePath' make functest
```                                                                                       
