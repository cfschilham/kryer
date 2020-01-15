GO := go
VERSION := v1.2.1

.PHONY: build release release_windows release_linux release_darwin clean test deps

build:
	$(GO) build -v

release:
	mkdir autossh-$(VERSION)
	cp -r cfg autossh-$(VERSION)
	cp LICENSE autossh-$(VERSION)/LICENSE
	$(GO) build -v -o autossh-$(VERSION)/

release_windows:
	mkdir autossh-$(VERSION)-windows-amd64
	cp -r cfg autossh-$(VERSION)-windows-amd64
	cp LICENSE autossh-$(VERSION)-windows-amd64/LICENSE
	env GOOS=windows GOARCH=amd64 $(GO) build -v -o autossh-$(VERSION)-windows-amd64/

release_linux:
	mkdir autossh-$(VERSION)-linux-amd64
	cp -r cfg autossh-$(VERSION)-linux-amd64
	cp LICENSE autossh-$(VERSION)-linux-amd64/LICENSE
	env GOOS=linux GOARCH=amd64 $(GO) build -v -o autossh-$(VERSION)-linux-amd64/

release_darwin:
	mkdir autossh-$(VERSION)-darwin-amd64
	cp -r cfg autossh-$(VERSION)-darwin-amd64
	cp LICENSE autossh-$(VERSION)-darwin-amd64/LICENSE
	env GOOS=darwin GOARCH=amd64 $(GO) build -v -o autossh-$(VERSION)-darwin-amd64/

clean:
	$(GO) clean
	rm -rf autossh-$(VERSION)
	rm -rf autossh-$(VERSION)-windows-amd64
	rm -rf autossh-$(VERSION)-linux-amd64
	rm -rf autossh-$(VERSION)-darwin-amd64

test:
	$(GO) test ./...

deps:
	$(GO) get -v -u golang.org/x/crypto/ssh
	$(GO) get -v -u github.com/spf13/viper
	$(GO) get -v -u github.com/fatih/color
