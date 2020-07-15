Updating image list for pre pulling
-----------------------------------

`fetch-images.sh` can be called to extract the image identifiers from the manifests and the shell scripts in the provision dir that exists per k8s version.

However, it does **not** retrieve transitive dependencies. Therefore it can occur that during cluster up check there are images found that do not appear in the list. You can then just manually add them to the file `extra-pre-pull-images` and therefore achieve that transitive images are then also pre pulled during provisioning.

Example:

initially no extra-pre-pull-images present

During cluster-up check error:
```bash
...
Images found in cluster that are not in list!
fluent/fluentd:v1.2-debian
```

After adding the above image to the file `fluent/fluentd:v1.2-debian` and reprovisioning the error should go away.

Checking pull policies
----------------------

There should not be any pull policies present that lead to always pulling the images (see [here](https://kubernetes.io/docs/concepts/containers/images/#updating-images)) as that would avoid using the already pre pulled images. There are two ways to check:
* use [`./cluster-up/up.sh`](../../cluster-up/up.sh) and call script [`validate-pod-pull-policies.sh`](validate-pod-pull-policies.sh), which will check the policies on the deployed pods
* call script [`validate-manifest-pull-policies.sh`](validate-manifest-pull-policies.sh) which will check the policies on the manifests

The check on pods is already performed during cluster-up check after provisioning the cluster. 
TODO: it's not causing the build to fail (yet).