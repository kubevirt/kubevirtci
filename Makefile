cluster-up:
	./cluster-up/check.sh
	./cluster-up/up.sh

cluster-down:
	./cluster-up/down.sh

connect:
	@./cluster-up/container.sh

scale-up:
	./cluster-up/scale-up.sh

.PHONY: \
	cluster-up \
	cluster-down 
