all: build

build: mrun

mrun: FORCE
	CGO_ENABLED=0 go build -o $@ .

check: vet

vet:
	go vet $(CURDIR)/...

test:
	go test -race -count 1 $(CURDIR)/...

FORCE:

.PHONY: all build check vet test
