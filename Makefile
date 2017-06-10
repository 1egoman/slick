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
	for i in `find . -type d ! -path "./vendor*" ! -path "./.git*"`; do \
		go fmt $i; \
	done
	git add -u
	git commit -m "chore(fmt): ran go fmt on code"

# ===
# Tests and coverage
# ===

test:
	go test -v ./...

test-cov:
	mkdir -p coverage/
	go test -c -covermode=count -coverpkg ./... 
	./slick.test -test.coverprofile coverage.cov
	mv coverage.cov coverage/
	go tool cover -html=./coverage/coverage.cov -o coverage/index.html

.PHONY: clean test test-cov
