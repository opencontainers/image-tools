GO15VENDOREXPERIMENT=1
export GO15VENDOREXPERIMENT
PREFIX ?= $(DESTDIR)/usr
BINDIR ?= $(DESTDIR)/usr/bin


default: all

help:
	@echo "Usage: make <target>"
	@echo
	@echo " * 'all' - Build the oci tool and manual pages"
	@echo " * 'tool' - Build the oci tool"
	@echo " * 'install' - Install binaries and manual pages"
	@echo " * 'install.tools' - Install tool needed for building this project"
	@echo " * 'uninstall' - Remove the oci tool and manual pages"
	@echo " * 'man' - Build the oci image manual pages"
	@echo " * 'check-license' - Check license headers in source files"
	@echo " * 'lint' - Execute the source code linter"
	@echo " * 'test' - Execute the unit tests"
	@echo " * 'update-deps' - Update vendored dependencies"
	@echo " * 'clean' - clean existing build artifacts"

check-license:
	@echo "checking license headers"
	@./.tool/check-license

.PHONY: tool
tool:
	go build -o oci-image-tool ./cmd/oci-image-tool


all: tool man

.PHONY: man
man:
	go-md2man -in "man/oci-image-tool.1.md" -out "oci-image-tool.1"
	go-md2man -in "man/oci-image-tool-create.1.md" -out "oci-image-tool-create.1"
	go-md2man -in "man/oci-image-tool-unpack.1.md" -out "oci-image-tool-unpack.1"
	go-md2man -in "man/oci-image-tool-validate.1.md" -out "oci-image-tool-validate.1"


install: man
	install -d -m 755 $(BINDIR)
	install -m 755 oci-image-tool $(BINDIR)
	install -d -m 755 $(PREFIX)/share/man/man1
	install -m 644 *.1 $(PREFIX)/share/man/man1
	install -d -m 755 $(PREFIX)/share/bash-completion/completions
	install -m 644 completions/bash/oci-image-tool $(PREFIX)/share/bash-completion/completionsn

uninstall:
	rm -f $(BINDIR)/oci-image-tool
	rm -f $(PREFIX)/share/man/man1/oci-image-tool*.1
	rm -f $(PREFIX)/share/bash-completion/completions/oci-image-tool

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

install.tools: .install.gitvalidation .install.glide .install.glide-vc .install.gometalinter .install.go-md2man

.install.gitvalidation:
	go get github.com/vbatts/git-validation

.install.glide:
	go get github.com/Masterminds/glide

.install.glide-vc:
	go get github.com/sgotti/glide-vc

.install.gometalinter:
	go get github.com/alecthomas/gometalinter
	gometalinter --install --update

.install.go-md2man:
	go get github.com/cpuguy83/go-md2man

clean:
	rm -rf oci-image-tool *.1

.PHONY: \
	all \
	tool \
	man \
	install \
	uninstall \
	check-license \
	clean \
	lint \
	test
