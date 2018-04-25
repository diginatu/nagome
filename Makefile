GOFILES = $(shell find . -name '*.go' -not -path './vendor/*')
GOPACKAGES = $(shell go list ./...  | grep -v /vendor/)
BINARY = nagome
RELEASEDIR = release
LDFLAGS = -ldflags "-X main.Version=`git describe --tags`"

default: build

.PHONY: build
build: $(GOFILES)
	go build $(LDFLAGS) -o $(BINARY)

.PHONY: install
install:
	go install $(LDFLAGS)

.PHONY: cross
cross:
	gox -osarch "linux/amd64 darwin/amd64 windows/amd64" -output "$(RELEASEDIR)/{{.Dir}}_{{.OS}}_{{.Arch}}" $(LDFLAGS)

.PHONY: clean
clean:
	if [ -f $(BINARY) ] ; then rm $(BINARY) ; fi
	rm -rf release

.PHONY: test
test:
	go test -v -race $(GOPACKAGES)

.PHONY: coverage
coverage:
	echo "" > coverage.txt
	for d in $(GOPACKAGES); do \
		go test -coverprofile=profile.out -covermode=atomic "$$d"; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi \
	done
