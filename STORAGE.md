## Storage providers inside the generated images.
In order to make the generated images more useful they now include a local volume storage provider by default. The StorageClass associated with the provider is set as the default storage class. So if your PVC request doesn't include a StorageClass, the local volume provider will be used to satisfy the request. local volume storage provider is no longer alpha since Kubernetes 1.10, so images based on 1.10 or newer will have the local volume provider enabled. Same thing goes for the OpenShift images, version 3.10 or newer will have them enabled.

By default, we have pre-provisioned 10 local volumes per node, which are available to make requests against. You can manually bind mount more directories to increase this number.

## Example PVC request
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: "pvc"
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 20Mi
```
This request doesn't specify the StorageClass, and thus will be provisioned by the 'default' Storage Class which is the local volume provisioner. To manually request the local volume provisioner, add the 'local' StorageClass to the PVC request.
