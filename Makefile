VERSION?=$(shell git describe --tags --dirty | sed 's/^v//')
GO_BUILD=CGO_ENABLED=0 go build -i --ldflags="-w"

rwildcard=$(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) \
    $(filter $(subst *,%,$2),$d))

LINTERS=\
	gofmt \
	golint \
	vet \
	misspell \
	ineffassign \
	deadcode

all: ci

ci: $(LINTERS) cover build

.PHONY: all ci

# ################################################
# Bootstrapping for base golang package deps
# ################################################

CMD_PKGS=\
	github.com/golang/lint/golint \
	honnef.co/go/simple/cmd/gosimple \
	github.com/client9/misspell/cmd/misspell \
	github.com/gordonklaus/ineffassign \
	github.com/tsenart/deadcode \
	github.com/alecthomas/gometalinter

define VENDOR_BIN_TMPL
vendor/bin/$(notdir $(1)): vendor
	go build -o $$@ ./vendor/$(1)
VENDOR_BINS += vendor/bin/$(notdir $(1))
endef

$(foreach cmd_pkg,$(CMD_PKGS),$(eval $(call VENDOR_BIN_TMPL,$(cmd_pkg))))
$(patsubst %,%-bin,$(filter-out gofmt vet,$(LINTERS))): %-bin: vendor/bin/%
gofmt-bin vet-bin:

bootstrap:
	glide -v || curl http://glide.sh/get | sh

vendor: glide.lock
	glide install

.PHONY: bootstrap $(CMD_PKGS)

# ################################################
# Test and linting
# ###############################################

test: vendor
	@CGO_ENABLED=0 go test -v $$(glide nv)

COVER_TEST_PKGS:=$(shell find . -type f -name '*_test.go' | grep -v vendor | rev | cut -d "/" -f 2- | rev | sort -u)
$(COVER_TEST_PKGS:=-cover): %-cover: all-cover.txt
	@CGO_ENABLED=0 go test -coverprofile=$@.out -covermode=atomic ./$*
	@if [ -f $@.out ]; then \
	    grep -v "mode: atomic" < $@.out >> all-cover.txt; \
	    rm $@.out; \
	fi

all-cover.txt:
	echo "mode: atomic" > all-cover.txt

cover: vendor all-cover.txt $(COVER_TEST_PKGS:=-cover)

$(LINTERS): %: vendor/bin/gometalinter %-bin vendor
	PATH=`pwd`/vendor/bin:$$PATH gometalinter --tests --disable-all --vendor \
	    --deadline=5m -s data $$(glide nv) --enable $@


.PHONY: cover $(LINTERS) $(COVER_TEST_PKGS:=-cover)

# ################################################
# Building
# ###############################################$

PREFIX?=
SUFFIX=
ifeq ($(GOOS),windows)
    SUFFIX=.exe
endif

build: $(PREFIX)bin/manifold-cli

MANIFOLDCLI_DEPS=\
		vendor \
		$(wildcard *.go) \
		$(call rwildcard,cmd,*.go) \

$(PREFIX)bin/manifold-cli$(SUFFIX): $(MANIFOLDCLI_DEPS)
	$(GO_BUILD) -o $(PREFIX)bin/manifold-cli$(SUFFIX) ./cmd

.PHONY: build

# ################################################
# Cleaning
# ################################################

clean:
	rm -rf bin/manifold-cli
	rm -rf bin/manifold-cli.exe
	rm -rf build
