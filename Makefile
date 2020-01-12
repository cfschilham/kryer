GO := go
VERSION := v1.1.4

.PHONY: build windows linux darwin clean test deps

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
	env GOOS=windows GOARCH=amd64 $(GO) build -v -o autossh-$(VERSION)-windows-amd64/$(BUILD_TARGET).exe

release_linux:
	mkdir autossh-$(VERSION)-linux-amd64
	cp -r cfg autossh-$(VERSION)-linux-amd64
	cp LICENSE autossh-$(VERSION)-linux-amd64/LICENSE
	env GOOS=linux GOARCH=amd64 $(GO) build -v -o autossh-$(VERSION)-linux-amd64/$(BUILD_TARGET)

release_darwin:
	mkdir autossh-$(VERSION)-darwin-amd64
	cp -r cfg autossh-$(VERSION)-darwin-amd64
	cp LICENSE autossh-$(VERSION)-darwin-amd64/LICENSE
	env GOOS=darwin GOARCH=amd64 $(GO) build -v -o autossh-$(VERSION)-darwin-amd64/$(BUILD_TARGET)

clean:
	$(GO) clean
	rm -r autossh-$(VERSION)
	rm -r autossh-$(VERSION)-windows-amd64
	rm -r autossh-$(VERSION)-linux-amd64
	rm -r autossh-$(VERSION)-darwin-amd64

test:
	$(GO) test ./...

deps:
	$(GO) get -v golang.org/x/crypto/ssh
	$(GO) get -v github.com/spf13/viper
