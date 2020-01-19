package loadcfg

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileToSlice(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "autossh-test-*.txt")
	if err != nil {
		t.Errorf("failed to create temp file: %s", err.Error())
		return
	}

	path := f.Name()
	defer f.Close()
	defer os.Remove(path)

	if _, err := f.WriteString("line1\nline2\nline3\nline4\nline5\n"); err != nil {
		t.Errorf("failed to write to temp file: %s", err.Error())
		return
	}

	want := []string{"line1", "line2", "line3", "line4", "line5"}
	got, err := fileToSlice(path)
	if err != nil {
		t.Errorf("failed to run fuction: %s", err.Error())
		return
	}

	for i := 0; i < len(got); i++ {
		if got[i] != want[i] {
			t.Errorf("incorrect result, want: %s got: %s", want, got)
			return
		}
	}
}

func TestLoadDict(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "autossh-test-*.txt")
	if err != nil {
		t.Errorf("failed to create temp file: %s", err.Error())
		return
	}

	path := f.Name()
	defer f.Close()
	defer os.Remove(path)

	if _, err := f.WriteString("line1\nline2\nline3\nline4\nline5\n"); err != nil {
		t.Errorf("failed to write to temp file: %s", err.Error())
		return
	}

	want := []string{"line1", "line2", "line3", "line4", "line5"}
	got, err := LoadDict(path)
	if err != nil {
		t.Errorf("failed to run fuction: %s", err.Error())
		return
	}

	for i := 0; i < len(got.Pwds()); i++ {
		if got.Pwds()[i] != want[i] {
			t.Errorf("incorrect result, want: %s got: %s", want, got.Pwds())
			return
		}
	}
}

func TestStrToHost(t *testing.T) {
	type params struct {
		str       string
		usrIsHost bool
	}
	testCases := []struct {
		params params
		want   Host
	}{
		{
			params: params{str: "user@host.local", usrIsHost: false},
			want:   Host{username: "user", addr: "host.local"},
		},
		{
			params: params{str: "user", usrIsHost: true},
			want:   Host{username: "user", addr: "user.local"},
		},
		{
			params: params{str: "user123@192.168.0.0", usrIsHost: false},
			want:   Host{username: "user123", addr: "192.168.0.0"},
		},
		{
			params: params{str: "username", usrIsHost: true},
			want:   Host{username: "username", addr: "username.local"},
		},
		{
			params: params{str: "root@localhost", usrIsHost: false},
			want:   Host{username: "root", addr: "localhost"},
		},
		{
			params: params{str: "username321", usrIsHost: true},
			want:   Host{username: "username321", addr: "username321.local"},
		},
	}

	for _, testCase := range testCases {
		got, err := StrToHost(testCase.params.str, testCase.params.usrIsHost)
		if err != nil {
			t.Errorf("failed to run function: %s", err.Error())
			return
		}

		if got != testCase.want {
			t.Errorf("incorrect result, want: %s got: %s", testCase.want, got)
			return
		}
	}
}

func TestSlcToHosts(t *testing.T) {
	type params struct {
		slc       []string
		usrIsHost bool
	}
	testCases := []struct {
		params params
		want   []Host
	}{
		{
			params: params{slc: []string{"user", "user2", "user3"}, usrIsHost: true},
			want: []Host{
				Host{username: "user", addr: "user.local"},
				Host{username: "user2", addr: "user2.local"},
				Host{username: "user3", addr: "user3.local"},
			},
		},
		{
			params: params{slc: []string{"user@192.168.0.0", "user2@192.168.0.0", "user3@192.168.0.0"}, usrIsHost: false},
			want: []Host{
				Host{username: "user", addr: "192.168.0.0"},
				Host{username: "user2", addr: "192.168.0.0"},
				Host{username: "user3", addr: "192.168.0.0"},
			},
		},
	}

	for _, testCase := range testCases {
		got, err := SlcToHosts(testCase.params.slc, testCase.params.usrIsHost)
		if err != nil {
			t.Errorf("failed to run function: %s", err.Error())
			return
		}

		if len(got) != len(testCase.want) {
			t.Errorf("discrepancy between amount of results, want: %d got: %d", len(testCase.want), len(got))
			return
		}
		for i := 0; i < len(testCase.want); i++ {
			if got[i] != testCase.want[i] {
				t.Errorf("incorrect result, want: %s got %s", testCase.want, got)
			}
		}
	}
}
