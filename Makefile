GO := go
VERSION := v2.0.0

.PHONY: build release release_windows release_linux release_darwin clean test deps

build:
	$(GO) build -v

install:
	@if [ ! -d /usr/bin ]; then \
		echo platform unsupported; \
		exit 1; \
	fi

	@if [ ! -w /usr/bin ]; then \
		echo insufficient permissions, please elevate; \
		exit 1; \
	fi

	@if [ ! -f kryer ]; then \
		$(MAKE) build; \
	fi

	@cp kryer /usr/bin/kryer && echo installed successfully, use command kryer to start

release_windows:
	mkdir kryer-$(VERSION)-windows-amd64
	env GOOS=windows GOARCH=amd64 $(GO) build -v -o kryer-$(VERSION)-windows-amd64/

release_linux:
	mkdir kryer-$(VERSION)-linux-amd64
	env GOOS=linux GOARCH=amd64 $(GO) build -v -o kryer-$(VERSION)-linux-amd64/

release_darwin:
	mkdir kryer-$(VERSION)-darwin-amd64
	env GOOS=darwin GOARCH=amd64 $(GO) build -v -o kryer-$(VERSION)-darwin-amd64/

clean:
	$(GO) clean
	rm -rf kryer-$(VERSION)
	rm -rf kryer-$(VERSION)-windows-amd64
	rm -rf kryer-$(VERSION)-linux-amd64
	rm -rf kryer-$(VERSION)-darwin-amd64

test:
	$(GO) test ./...

deps:
	$(GO) get -v -u golang.org/x/crypto/ssh
	$(GO) get -v -u github.com/spf13/viper
	$(GO) get -v -u github.com/fatih/color
