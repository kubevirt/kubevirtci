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