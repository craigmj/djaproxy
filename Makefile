SHELL=/bin/bash
GOSRC=$(shell find . -name "*.go")

bin/djaproxy: $(GOSRC)
	go build -o bin/djaproxy cmd/djaproxy.go
