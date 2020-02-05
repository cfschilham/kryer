GO := go
VERSION := v2.0.1
INSTALL_DIR := /usr/bin

ifeq ($(shell uname), Darwin)
	INSTALL_DIR = /usr/local/bin
endif

.PHONY: build release_windows release_linux release_darwin clean test deps

build:
	$(GO) build -v

install:
	@if [ ! -d $(INSTALL_DIR) ]; then \
		echo "Unable to locate $(INSTALL_DIR)"; \
		exit 1; \
	fi

	@if [ ! -w $(INSTALL_DIR) ]; then \
		echo "Insufficient permissions, please elevate"; \
		exit 1; \
	fi

	@if [ ! -f kryer ]; then \
		$(MAKE) build; \
	fi

	@cp kryer $(INSTALL_DIR)/kryer && echo "Installed successfully into $(INSTALL_DIR), use kryer command to start"; \

release_windows:
	mkdir kryer-$(VERSION)-windows-amd64
	env GOOS=windows GOARCH=amd64 $(GO) build -v -o kryer-$(VERSION)-windows-amd64/
	zip -9 kryer-$(VERSION)-windows-amd64.zip kryer-$(VERSION)-windows-amd64/kryer.exe
	echo `shasum -a 256 kryer-$(VERSION)-windows-amd64.zip` > kryer-$(VERSION)-windows-amd64.sha256
	rm -rf kryer-$(VERSION)-windows-amd64

release_linux:
	mkdir kryer-$(VERSION)-linux-amd64
	env GOOS=linux GOARCH=amd64 $(GO) build -v -o kryer-$(VERSION)-linux-amd64/
	tar -cvzf kryer-$(VERSION)-linux-amd64.tar.gz kryer-$(VERSION)-linux-amd64/kryer
	echo `shasum -a 256 kryer-$(VERSION)-linux-amd64.tar.gz` > kryer-$(VERSION)-linux-amd64.sha256
	rm -rf kryer-$(VERSION)-linux-amd64

release_darwin:
	mkdir kryer-$(VERSION)-darwin-amd64
	env GOOS=darwin GOARCH=amd64 $(GO) build -v -o kryer-$(VERSION)-darwin-amd64/
	tar -cvzf kryer-$(VERSION)-darwin-amd64.tar.gz kryer-$(VERSION)-darwin-amd64/kryer
	echo `shasum -a 256 kryer-$(VERSION)-darwin-amd64.tar.gz` > kryer-$(VERSION)-darwin-amd64.sha256
	rm -rf kryer-$(VERSION)-darwin-amd64

clean:
	$(GO) clean
	rm -rf kryer-$(VERSION)
	rm -rf kryer-$(VERSION)-windows-amd64
	rm -rf kryer-$(VERSION)-linux-amd64
	rm -rf kryer-$(VERSION)-darwin-amd64
	rm -f kryer-$(VERSION)-windows-amd64.zip
	rm -f kryer-$(VERSION)-linux-amd64.tar.gz
	rm -f kryer-$(VERSION)-darwin-amd64.tar.gz
	rm -f kryer-$(VERSION)-windows-amd64.sha256
	rm -f kryer-$(VERSION)-linux-amd64.sha256
	rm -f kryer-$(VERSION)-darwin-amd64.sha256

test:
	$(GO) test ./...

deps:
	$(GO) get -v -u golang.org/x/crypto/ssh
	$(GO) get -v -u github.com/fatih/color