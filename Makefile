VERSION?=$(shell git describe --tags --dirty | sed 's/^v//')
PKG=github.com/manifoldco/manifold-cli
STRIPE_PKEY=${STRIPE_PUBLISHABLE_KEY}
GO_BUILD=CGO_ENABLED=0 go build -i --ldflags="-w -X $(PKG)/config.Version=$(VERSION) -X $(PKG)/config.StripePublishableKey=$(STRIPE_PKEY) -X $(PKG)/config.GitHubClientID=$(GITHUB_CLIENT_ID)"

PROMULGATE_VERSION=0.0.8

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

ci: generated-clients $(LINTERS) cover build

.PHONY: all ci

# ################################################
# Bootstrapping for base golang package deps
# ################################################

CMD_PKGS=\
	github.com/golang/lint/golint \
	github.com/client9/misspell/cmd/misspell \
	github.com/gordonklaus/ineffassign \
	github.com/tsenart/deadcode \
	github.com/alecthomas/gometalinter \
	github.com/go-swagger/go-swagger/cmd/swagger

define VENDOR_BIN_TMPL
vendor/bin/$(notdir $(1)): vendor/$(1) | vendor
	go build -o $$@ ./vendor/$(1)
VENDOR_BINS += vendor/bin/$(notdir $(1))
vendor/$(1): Gopkg.lock
	dep ensure -vendor-only
endef

$(foreach cmd_pkg,$(CMD_PKGS),$(eval $(call VENDOR_BIN_TMPL,$(cmd_pkg))))

$(patsubst %,%-bin,$(filter-out gofmt vet,$(LINTERS))): %-bin: vendor/bin/%
gofmt-bin vet-bin:

bootstrap:
	which dep || go get github.com/golang/dep/cmd/dep

vendor: Gopkg.lock
	dep ensure

.PHONY: bootstrap $(CMD_PKGS)

# ################################################
# Test and linting
# ###############################################

test: vendor
	@CGO_ENABLED=0 go test -v $$(go list ./... | grep -v vendor)

COVER_TEST_PKGS:=$(shell find . -type f -name '*_test.go' | grep -v vendor | grep -v generated | rev | cut -d "/" -f 2- | rev | sort -u)
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
	     --deadline=5m -s data --skip generated --enable $@

.PHONY: cover $(LINTERS) $(COVER_TEST_PKGS:=-cover)

# ################################################
# Building Swagger Clients
# ###############################################

generated/%/client: specs/%.yaml vendor/bin/swagger
	vendor/bin/swagger generate client -f $< -t generated/$*
	touch generated/$*/client
	touch generated/$*/models

APIS=$(patsubst specs/%.yaml,%,$(wildcard specs/*.yaml))
API_CLIENTS=$(APIS:%=generated/%/client)
generated-clients: $(API_CLIENTS)

.PHONY: generated-clients

# ################################################
# Building
# ###############################################$

PREFIX?=
SUFFIX=
ifeq ($(GOOS),windows)
    SUFFIX=.exe
endif

build: $(PREFIX)bin/manifold$(SUFFIX)

MANIFOLDCLI_DEPS=\
		vendor \
		$(wildcard *.go) \
		$(call rwildcard,cmd,*.go) \
		generated-clients

$(PREFIX)bin/manifold$(SUFFIX): $(MANIFOLDCLI_DEPS)
	$(GO_BUILD) -o $(PREFIX)bin/manifold$(SUFFIX) ./cmd

.PHONY: build

#################################################
# Releasing
#################################################

NO_WINDOWS= \
	darwin_amd64 \
	linux_amd64
OS_ARCH= \
	$(NO_WINDOWS) \
	windows_amd64

os=$(word 1,$(subst _, ,$1))
arch=$(word 2,$(subst _, ,$1))

os-build/windows_amd64/bin/manifold: os-build/%/bin/manifold:
	PREFIX=build/$*/ GOOS=$(call os,$*) GOARCH=$(call arch,$*) make build/$*/bin/manifold.exe
$(NO_WINDOWS:%=os-build/%/bin/manifold): os-build/%/bin/manifold:
	PREFIX=build/$*/ GOOS=$(call os,$*) GOARCH=$(call arch,$*) make build/$*/bin/manifold

build/manifold-cli_$(VERSION)_windows_amd64.zip: build/manifold-cli_$(VERSION)_%.zip: os-build/%/bin/manifold
	cd build/$*/bin; zip -r ../../manifold-cli_$(VERSION)_$*.zip manifold.exe
$(NO_WINDOWS:%=build/manifold-cli_$(VERSION)_%.tar.gz): build/manifold-cli_$(VERSION)_%.tar.gz: os-build/%/bin/manifold
	cd build/$*/bin; tar -czf ../../manifold-cli_$(VERSION)_$*.tar.gz manifold

zips: $(NO_WINDOWS:%=build/manifold-cli_$(VERSION)_%.tar.gz) build/manifold-cli_$(VERSION)_windows_amd64.zip

release: zips
	curl -LO https://releases.manifold.co/promulgate/$(PROMULGATE_VERSION)/promulgate_$(PROMULGATE_VERSION)_linux_amd64.tar.gz
	tar xvf promulgate_*
	./promulgate release v$(VERSION)

.PHONY: release zips $(OS_ARCH:%=os-build/%/bin/manifold)

# ################################################
# Cleaning
# ################################################

clean:
	rm -rf bin/manifold
	rm -rf bin/manifold.exe
	rm -rf build
	rm -rf generated
