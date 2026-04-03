run: build
	./pjf

uninstall: build
	./pjf uninstall

build:
	go build -o pjf .

.PHONY: run uninstall build
