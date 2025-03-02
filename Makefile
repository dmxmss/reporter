TOKEN := $(shell cat .gittoken)
BOTTOKEN := $(shell cat .bottoken)

run:
	go build main.go
	mv main build
	TOKEN=$(TOKEN) BOTTOKEN=$(BOTTOKEN) ./build/main
r:
	TOKEN=$(TOKEN) BOTTOKEN=$(BOTTOKEN) ./build/main
b:
	go build src/main.go
	mv main build
