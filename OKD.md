# [kubevirtci](README.md): Getting Started with multi-node OKD Provider

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
                                                                                                                                            
## OKD Console                                                                                                                              
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
