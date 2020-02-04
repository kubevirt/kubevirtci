# How to create new OCP release

1. Get pull secrets from https://cloud.redhat.com/openshift/install/metal/user-provisioned

2. Export location INSTALLER_PULL_SECRET=pull-secret.txt

2. Provision ocp-cnao-4.4 provider ./cluster-provision/ocp/4.4/provision.sh

3. Log into quay.io container registry make sure you have push permissiong for openshift-cnv organization

4. Push the ocp-cnao-4.4 provider container with ./cluster-provision/ocp-cnao/4.4/publish.sh
