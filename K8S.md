# [kubevirtci](README.md): Getting Started with a multi-node Kubernetes Provider                              
                                                                                      
Download this repo                                                                    
```                                                                                   
git clone https://github.com/kubevirt/kubevirtci.git                                  
cd kubevirtci                                                                         
```                                                                                   
                                                                                      
Start multi node k8s cluster with 2 nics                                              
```                                                                                   
export KUBEVIRT_PROVIDER=k8s-1.22 KUBEVIRT_NUM_NODES=2 KUBEVIRT_NUM_SECONDARY_NICS=1
make cluster-up                                                                       
```                                                                                   
                                                                                      
Stop k8s cluster                                                                      
```
make cluster-down                                                                     
```

Use provider's kubectl client with kubectl.sh wrapper script               
```
export KUBEVIRTCI_TAG=`curl -L -Ss https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirtci/latest`
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
                                                                           
Attach to node console with screen and pty
```                                                  
# Attach to node01 console                           
docker exec -it ${KUBEVIRT_PROVIDER}-node01 screen /dev/pts/0
```                                                 
