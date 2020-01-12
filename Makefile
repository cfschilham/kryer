GO := go
BUILD_TARGET := autossh
VERSION := v1.1.4

.PHONY: build windows linux darwin clean test deps

build:
	$(GO) build -v -o $(BUILD_TARGET)

windows:
	mkdir autossh-$(VERSION)-windows-amd64
	cp -r cfg autossh-$(VERSION)-windows-amd64
	cp LICENSE autossh-$(VERSION)-windows-amd64/LICENSE
	env GOOS=windows GOARCH=amd64 $(GO) build -v -o autossh-$(VERSION)-windows-amd64/$(BUILD_TARGET).exe

linux:
	mkdir autossh-$(VERSION)-linux-amd64
	cp -r cfg autossh-$(VERSION)-linux-amd64
	cp LICENSE autossh-$(VERSION)-linux-amd64/LICENSE
	env GOOS=linux GOARCH=amd64 $(GO) build -v -o autossh-$(VERSION)-linux-amd64/$(BUILD_TARGET)

darwin:
	mkdir autossh-$(VERSION)-darwin-amd64
	cp -r cfg autossh-$(VERSION)-darwin-amd64
	cp LICENSE autossh-$(VERSION)-darwin-amd64/LICENSE
	env GOOS=darwin GOARCH=amd64 $(GO) build -v -o autossh-$(VERSION)-darwin-amd64/$(BUILD_TARGET)

clean:
	$(GO) clean

test:
	$(GO) test ./...

deps:
	$(GO) get -v golang.org/x/crypto/ssh
	$(GO) get -v github.com/spf13/viper
