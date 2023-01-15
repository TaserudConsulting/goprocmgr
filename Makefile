build:
	go build -o goprocmgr

# Just a make target to print numbers in a loop forever, it's nice to
# have a command to test with that generate logs.
printloop:
	echo "PORT: "$$PORT
	echo "PATH: "$$PATH
	num=0; while true; do \
		echo $$num; \
		sleep 5; \
		((num = num + 1)); \
	done
