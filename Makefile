
export KUBEVIRTCI_TAG ?= $(shell curl -L -Ss https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirtci/latest)

cluster-up:
	./cluster-up/check.sh
	./cluster-up/up.sh

cluster-down:
	./cluster-up/down.sh

bump-net-resources-injector:
	./hack/bump-net-resources-injector.sh

crio-verify:
	./hack/verify-crio-sync.sh

crio-sync:
	./hack/sync-crio-versions.sh

.PHONY: \
	cluster-up \
	cluster-down \
	bump-net-resources-injector \
	bump \
	crio-verify \
	crio-sync
