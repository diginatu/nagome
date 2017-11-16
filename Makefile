GOFILES = $(shell find . -name '*.go' -not -path './vendor/*')
GOPACKAGES = $(shell go list ./...  | grep -v /vendor/)
BINARY = nagome
RELEASEDIR = release
LDFLAGS = -ldflags "-X main.Version=`git describe`"

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
