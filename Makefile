build:
	go build -o goprocmgr

# Just a make target to print numbers in a loop forever, it's nice to
# have a command to test with that generate logs.
printloop:
	@echo "PORT: "$$PORT
	@echo "PATH: "$$PATH
	@num=0; while true; do \
		echo "[`date +'%Y-%m-%d %H:%M:%S'`] stdout: $$num"; \
		sleep 3; \
		echo "[`date +'%Y-%m-%d %H:%M:%S'`] stderr: $$num" >&2; \
		sleep 3; \
		((num = num + 1)); \
	done
