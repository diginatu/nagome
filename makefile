SOURCES := $(shell find . -name '*.go')
BINARY=nagome
LDFLAGS=-ldflags "-X main.Version=`git describe`"

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES)
	go build ${LDFLAGS} -o ${BINARY}

.PHONY: install
install:
	go install ${LDFLAGS}

.PHONY: cross
cross:
	gox -osarch "linux/amd64 darwin/amd64 windows/amd64" -output "release/{{.Dir}}_{{.OS}}_{{.Arch}}" ${LDFLAGS}

.PHONY: clean
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
	rm -rf release

