package loadcfg

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileToSlice(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "autossh-test-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %s", err.Error())
	}

	path := f.Name()
	defer f.Close()
	defer os.Remove(path)

	if _, err := f.WriteString("line1\nline2\nline3\nline4\nline5\n"); err != nil {
		t.Fatalf("failed to write to temp file: %s", err.Error())
	}

	got, err := fileToSlice(path)
	if err != nil {
		t.Fatalf("failed to run fuction: %s", err.Error())
	}

	want := []string{"line1", "line2", "line3", "line4", "line5"}

	for i, entry := range got {
		if entry != want[i] {
			t.Fatalf("incorrect result, want: %s got: %s", want, got)
		}
	}
}

func TestLoadDict(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "autossh-test-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %s", err.Error())
	}

	path := f.Name()
	defer f.Close()
	defer os.Remove(path)

	if _, err := f.WriteString("line1\nline2\nline3\nline4\nline5\n"); err != nil {
		t.Fatalf("failed to write to temp file: %s", err.Error())
	}

	got, err := LoadDict(path)
	if err != nil {
		t.Fatalf("failed to run fuction: %s", err.Error())
	}

	want := []string{"line1", "line2", "line3", "line4", "line5"}

	for i, entry := range got.Pwds() {
		if entry != want[i] {
			t.Fatalf("incorrect result, want: %s got: %s", want, got.Pwds())
		}
	}
}

func TestStrToHost(t *testing.T) {
	tests := []struct {
		str       string
		usrIsHost bool
		host      Host
	}{
		{str: "user@host.local", usrIsHost: false, host: Host{username: "user", addr: "host.local"}},
		{str: "user", usrIsHost: true, host: Host{username: "user", addr: "user.local"}},
		{str: "user123@192.168.0.0", usrIsHost: false, host: Host{username: "user123", addr: "192.168.0.0"}},
		{str: "username", usrIsHost: true, host: Host{username: "username", addr: "username.local"}},
		{str: "root@localhost", usrIsHost: false, host: Host{username: "root", addr: "localhost"}},
		{str: "username321", usrIsHost: true, host: Host{username: "username321", addr: "username321.local"}},
	}

	for _, entry := range tests {
		host, err := StrToHost(entry.str, entry.usrIsHost)
		if err != nil {
			t.Fatalf("failed to run function")
		}

		if host != entry.host {
			t.Fatalf("incorrect result, want: %s got: %s", entry.host, host)
		}
	}
}

func TestSlcToHosts(t *testing.T) {
	tests := []struct {
		slc       []string
		usrIsHost bool
		hosts     []Host
	}{
		{
			slc:       []string{"user", "user2", "user3"},
			usrIsHost: true,
			hosts: []Host{
				Host{username: "user", addr: "user.local"},
				Host{username: "user2", addr: "user2.local"},
				Host{username: "user3", addr: "user3.local"},
			},
		},
		{
			slc:       []string{"user@192.168.0.0", "user2@192.168.0.0", "user3@192.168.0.0"},
			usrIsHost: false,
			hosts: []Host{
				Host{username: "user", addr: "192.168.0.0"},
				Host{username: "user2", addr: "192.168.0.0"},
				Host{username: "user3", addr: "192.168.0.0"},
			},
		},
	}

	for _, entry := range tests {
		hosts, err := SlcToHosts(entry.slc, entry.usrIsHost)
		if err != nil {
			t.Fatalf("failed to run function")
		}

		for i, host := range entry.hosts {
			if hosts[i] != host {
				t.Fatalf("incorrect result, want: %s got: %s", entry.hosts, hosts)
			}
		}
	}
}
