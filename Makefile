clean:
	rm -rf coverage/


# ===
# `go fmt`
# ===

is-clean:
	if [ -n "`git status -s | sed -n '/^ M/p'`" ]; then \
		echo "Error: Working tree not clean."; \
		exit 1; \
	fi

fmt: is-clean
	for i in `go list ./...`; do \
		go fmt $i; \
	done
	git add -u
	git commit -m "chore(fmt): ran go fmt on code"

# ===
# Tests and coverage
# ===

test:
	time for dir in `go list ./... | grep -v vendor`; do \
		go test $$dir; \
	done

test-cov:
	mkdir -p coverage/

	echo "mode: set" > coverage/acc.out
	time for dir in `go list ./... | grep -v vendor`; do \
		returnval=`go test -coverprofile=coverage/profile.out $$dir`; \
		echo $$returnval; \
		cat coverage/profile.out | grep -v "mode: set" >> coverage/acc.out; \
	done

	go tool cover -html=./coverage/acc.out -o ./coverage/index.html

.PHONY: clean test test-cov
