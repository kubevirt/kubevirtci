KUBEVIRTCI_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/../.. && pwd )"

okd_base_hash="sha256:73ede51ce464546a82b81956b7f58cf98662a4c5fded9c659b57746bc131e047"
gocli_image_hash="sha256:220f55f6b1bcb3975d535948d335bd0e6b6297149a3eba1a4c14cad9ac80f80d"

gocli="docker run \
--privileged \
--net=host \
--rm -t \
-v /var/run/docker.sock:/var/run/docker.sock \
-v ${KUBEVIRTCI_PATH}:${KUBEVIRTCI_PATH} \
docker.io/kubevirtci/gocli@${gocli_image_hash}"


