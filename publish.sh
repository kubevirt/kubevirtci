#!/bin/bash

set -ex

readonly ARTIFACTS=${ARTIFACTS:-${PWD}}
readonly CONCURRENT_PROVISION_JOBS_COUNT="${CONCURRENT_PROVISION_JOBS_COUNT:-2}"

KUBEVIRTCI_TAG=$(date +"%y%m%d%H%M")-$(git rev-parse --short HEAD)
export KUBEVIRTCI_TAG

TARGET_REPO="quay.io/kubevirtci"
TARGET_KUBEVIRT_REPO="quay.io/kubevirt"
TARGET_GIT_REMOTE="https://kubevirt-bot@github.com/kubevirt/kubevirtci.git"

# provision_on_background takes a function name and arguments list,
# executes the given command with each arg in the background and wait for it to finish.
# Usage:
# provision_on_background "foo" "a b c"
#   foo a &
#   foo b &
#   foo c &
# Each job stdout is exported to file at $ARTIFACTS, for example:
# foo-a.log
# foo-b.log
# foo-c.log
# The amount of background jobs is controlled by CONCURRENT_PROVISION_JOBS_COUNT env var.
exec_on_background() {
  local -r command=$1
  local -r args=($2)

  local -r jobs_count="${#args[@]}"
  local offset=0
  while [ "${jobs_count}" -gt "${offset}" ]; do
      jobs_batch_pids=()
      jobs_batch=("${args[@]:${offset}:${CONCURRENT_PROVISION_JOBS_COUNT}}")
      for job in ${jobs_batch[@]}; do
          job_name="${command}-${job}"
          eval "${command} ${job}" &> "${ARTIFACTS}/${job_name}.log" &
          jobs_batch_pids+=($!)
          echo "[$(date --utc)] ${job_name}, PID: $!.."
      done
      offset=$((offset+CONCURRENT_PROVISION_JOBS_COUNT))

      # wait for providers to finish
      set +e
      for job_pid in "${jobs_batch_pids[@]}"; do
          echo "[$(date --utc)] waiting for job, PID: ${job_pid}.."
          wait ${job_pid}
          result=$?
          echo "[$(date --utc)] job PID ${job_pid} finished with return code ${result}"
          if [ "${result}" -ne 0 ]; then
              echo "FATAL: base job failed, PID: ${job_pid}"
              exit 1
          fi
      done
      set -e
  done
}

# Build gocli
(cd cluster-provision/gocli && make container)
docker tag ${TARGET_REPO}/gocli ${TARGET_REPO}/gocli:${KUBEVIRTCI_TAG}

# Provision all base images
provision_cluster_node_base_image() {
    local -r os_name=$1
    echo "provisioning base: cd cluster-provision/${os_name} && ./build.sh"
    (cd cluster-provision/${os_name} && ./build.sh) &
}
readonly BASE_IMAGES=("centos8" "centos9")
exec_on_background "provision_cluster_node_base_image" "${BASE_IMAGES[*]}"

# provision all cluster image
provision_cluster_node_k8s_image() {
    local -r provider_name=$1

    local gocli_provision_flags=""
    if [[ $provider_name =~ ipv6 ]]; then
        gocli_provision_flags="--network-stack ipv6"
    else
        gocli_provision_flags="--network-stack dualstack"
    fi

    if [[ $provider_name =~ slim ]]; then
        gocli_provision_flags="${gocli_provision_flags} --slim"
    fi
    cluster-provision/gocli/build/cli provision cluster-provision/k8s/${provider_name} "${gocli_provision_flags}"
    docker tag ${TARGET_REPO}/k8s-${provider_name} ${TARGET_REPO}/k8s-${provider_name}:${KUBEVIRTCI_TAG}
}
readonly CLUSTERS=($(find cluster-provision/k8s/* -maxdepth 0 -type d -printf '%f\n'))
k8s_images=()
for c in "${CLUSTERS[@]}"; do
  k8s_images+=("${c}")
  k8s_images+=("${c}-slim")
done
exec_on_background "provision_cluster_node_k8s_image" "${k8s_images[*]}"

# Provision alpine container disk and tag it
(cd cluster-provision/images/vm-image-builder && ./create-containerdisk.sh alpine-cloud-init)
docker tag  alpine-cloud-init:devel ${TARGET_REPO}/alpine-with-test-tooling-container-disk:${KUBEVIRTCI_TAG}
docker tag  alpine-cloud-init:devel ${TARGET_KUBEVIRT_REPO}/alpine-with-test-tooling-container-disk:devel

# Push all images

# until "unknown blob" issue is fixed use skopeo to push image
# see https://github.com/moby/moby/issues/43234

IMAGES="${CLUSTERS}"
TARGET_IMAGE="${TARGET_REPO}/gocli:${KUBEVIRTCI_TAG}"
skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"
#docker push ${TARGET_IMAGE}
for i in ${IMAGES}; do
    TARGET_IMAGE="${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}"
    skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"
    #docker push ${TARGET_IMAGE}

    TARGET_IMAGE="${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}-slim"
    skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"
done

TARGET_IMAGE="${TARGET_REPO}/alpine-with-test-tooling-container-disk:${KUBEVIRTCI_TAG}"
skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"
#docker push ${TARGET_IMAGE}
TARGET_KUBEVIRT_IMAGE="${TARGET_KUBEVIRT_REPO}/alpine-with-test-tooling-container-disk:devel"
skopeo copy "docker-daemon:${TARGET_KUBEVIRT_IMAGE}" "docker://${TARGET_KUBEVIRT_IMAGE}"

git config user.name "kubevirt-bot"
git config user.email "rmohr+kubebot@redhat.com"
git tag ${KUBEVIRTCI_TAG}
git push ${TARGET_GIT_REMOTE} ${KUBEVIRTCI_TAG}
