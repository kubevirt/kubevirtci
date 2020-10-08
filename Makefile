cluster-up:
	./cluster-up/check.sh
	./cluster-up/up.sh

cluster-down:
	./cluster-up/down.sh

.PHONY: \
	cluster-up \
	cluster-down \
	bump
