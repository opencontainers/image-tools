GO15VENDOREXPERIMENT=1
export GO15VENDOREXPERIMENT

COMMIT=$(shell git rev-parse HEAD 2> /dev/null || true)

EPOCH_TEST_COMMIT ?= v0.2.0

default: help

help:
	@echo "Usage: make <target>"
	@echo
	@echo " * 'tools' - build the oci image tools binaries"
	@echo " * 'check-license' - check license headers in source files"
	@echo " * 'lint' - Execute the source code linter"
	@echo " * 'test' - Execute the unit tests"
	@echo " * 'update-deps' - Update vendored dependencies"

check-license:
	@echo "checking license headers"
	@./.tool/check-license

tools:
	go build -ldflags "-X main.gitCommit=${COMMIT}" ./cmd/oci-create-runtime-bundle
	go build -ldflags "-X main.gitCommit=${COMMIT}" ./cmd/oci-unpack
	go build -ldflags "-X main.gitCommit=${COMMIT}" ./cmd/oci-image-validate

lint:
	@echo "checking lint"
	@./.tool/lint

test:
	go test -race -cover $(shell go list ./... | grep -v /vendor/)

## this uses https://github.com/Masterminds/glide and https://github.com/sgotti/glide-vc
update-deps:
	@which glide > /dev/null 2>/dev/null || (echo "ERROR: glide not found. Consider 'make install.tools' target" && false)
	glide update --strip-vcs --strip-vendor --update-vendored --delete
	glide-vc --only-code --no-tests
	# see http://sed.sourceforge.net/sed1line.txt
	find vendor -type f -exec sed -i -e :a -e '/^\n*$$/{$$d;N;ba' -e '}' "{}" \;

.PHONY: .gitvalidation

# When this is running in travis, it will only check the travis commit range
.gitvalidation:
	@which git-validation > /dev/null 2>/dev/null || (echo "ERROR: git-validation not found. Consider 'make install.tools' target" && false)
ifeq ($(TRAVIS),true)
	git-validation -q -run DCO,short-subject,dangling-whitespace
else
	git-validation -v -run DCO,short-subject,dangling-whitespace -range $(EPOCH_TEST_COMMIT)..HEAD
endif

.PHONY: install.tools

install.tools: .install.gitvalidation .install.glide .install.glide-vc .install.gometalinter

.install.gitvalidation:
	go get github.com/vbatts/git-validation

.install.glide:
	go get github.com/Masterminds/glide

.install.glide-vc:
	go get github.com/sgotti/glide-vc

.install.gometalinter:
	go get github.com/alecthomas/gometalinter
	gometalinter --install --update

clean:
	rm -rf *~ $(OUTPUT_DIRNAME)
	rm -f oci-create-runtime-bundle
	rm -f oci-unpack
	rm -f oci-image-validate

.PHONY: \
	tools \
	check-license \
	clean \
	lint \
	test
