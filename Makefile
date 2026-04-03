run: build
	./pjf

uninstall: build
	./pjf uninstall

build:
	go build -o pjf .

test-repos:
	./scripts/create-test-repos.sh

clean-repos:
	./scripts/cleanup-test-repos.sh

.PHONY: run uninstall build test-repos clean-repos
