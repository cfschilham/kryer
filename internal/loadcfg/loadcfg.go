package loadcfg

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/spf13/viper"
)

// Config is used to load config values from cfg/config.yml
type Config struct {
	usrIsHost,
	multiThreaded bool
	goroutines int
	mode,
	port,
	dictPath,
	hostlistPath,
	outputPath string
}

// Dict is used to load a dictionary of passwords from the configured dictionary file, Dict.pwds
// represents the data stored in the dict file.
type Dict struct {
	pwds []string
}

// Host is used to represent a combination of a username and an ip to connect to via SSH.
type Host struct {
	username,
	addr string
}

// UsrIsHost returns the value of usrIsHost in a Config type.
func (c Config) UsrIsHost() bool {
	return c.usrIsHost
}

// MultiThreaded returns the value of multiThreaded in a Config type.
func (c Config) MultiThreaded() bool {
	return c.multiThreaded
}

// Goroutines returns the value of goroutines in a Config type.
func (c Config) Goroutines() int {
	return c.goroutines
}

// Mode returns the value of mode in a Config type.
func (c Config) Mode() string {
	return c.mode
}

// Port returns the value of port in a Config type.
func (c Config) Port() string {
	return c.port
}

// DictPath returns the value of dictPath in a Config type.
func (c Config) DictPath() string {
	return c.dictPath
}

// HostlistPath returns the value of hostlistPath in a Config type.
func (c Config) HostlistPath() string {
	return c.hostlistPath
}

// OutputPath returns the value of outputPath in a Config type.
func (c Config) OutputPath() string {
	return c.outputPath
}

// Pwds returns the value of pwds in a Dict type.
func (d *Dict) Pwds() []string {
	return d.pwds
}

// Addr returns the value of addr in a Host type.
func (h Host) Addr() string {
	return h.addr
}

// Username returns the value of username in a Host type.
func (h Host) Username() string {
	return h.username
}

// ResolveAddr tries to resolve the address using the cgo resolver.
func (h Host) ResolveAddr() (string, error) {
	resolver := net.Resolver{
		PreferGo: false,
	}

	addrs, err := resolver.LookupHost(context.Background(), h.addr)
	var ips []net.IP
	for _, addr := range addrs {
		ips = append(ips, net.ParseIP(addr))
	}

	if err != nil {
		return "", err
	}

	for _, ip := range ips {
		if ip.To4() != nil {
			return ip.To4().String(), nil
		}
	}
	return "", fmt.Errorf("internal/loadcfg: failed to resolve host: '%s'", h.addr)
}

// LoadConfig returns a config type based on the values in cfg/config.yml.
func LoadConfig(dir string) (*Config, error) {
	viper.SetConfigName("config.yml")
	viper.AddConfigPath(dir)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("internal/loadcfg: failed to load config: %s", err.Error())
	}

	c := &Config{
		usrIsHost:     viper.GetBool("user_is_host"),
		multiThreaded: viper.GetBool("multi_threaded"),
		goroutines:    viper.GetInt("goroutines"),
		mode:          viper.GetString("mode"),
		port:          viper.GetString("port"),
		dictPath:      viper.GetString("dict_path"),
		hostlistPath:  viper.GetString("hostlist_path"),
		outputPath:    viper.GetString("output_path"),
	}
	return c, nil
}

// fileToSlice opens a file at the given path and returns a string slice. For every line of the
// file a new entry is appended to the string slice.
func fileToSlice(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	slc := []string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		slc = append(slc, sc.Text())
	}
	if sc.Err() != nil {
		return nil, sc.Err()
	}

	return slc, nil
}

// LoadDict loads all entries from the file at the given path (designed for txt files) and returns
// them in the form of a Dictionary struct.
func LoadDict(path string) (*Dict, error) {
	slc, err := fileToSlice(path)
	if err != nil {
		return nil, fmt.Errorf("internal/loadcfg: failed to open %s: %s", path, err.Error())
	}
	return &Dict{pwds: slc}, nil
}

// StrToHost takes a string and returns a host. Strings should be passed in the form 'username@host'
// unless usrIsHost is true. With usrIsHost enabled, for example, an input of 'user1' means an output
// of a Host struct with username: 'user1', addr: 'user1.local'.
func StrToHost(str string, usrIsHost bool) (Host, error) {
	if str == "" {
		return Host{}, errors.New("internal/loadcfg: empty string passed")
	}

	if usrIsHost {
		return Host{
			username: str,
			addr:     str + ".local",
		}, nil
	}
	for i, char := range str {
		if string(char) == "@" && i != len(str)-1 {
			return Host{
				username: str[0:i],
				addr:     str[i+1:],
			}, nil
		}
	}
	return Host{}, fmt.Errorf("internal/loadcfg: invalid hostname '%s'", str)
}

// SlcToHosts takes a slice of strings and returns a slice of hosts. These hostscan be used to append
// to a Hostlist type. Strings should be passed in the form 'username@host' unless usrIsHost is true. With
// usrIsHost enabled, for example, an input of 'user1' means an output of a Host struct with username:
// 'user1', ip: 'user1.local'.
func SlcToHosts(slc []string, usrIsHost bool) ([]Host, error) {
	var hostSlc []Host
	for _, str := range slc {
		host, err := StrToHost(str, usrIsHost)
		if err != nil {
			return nil, err
		}
		hostSlc = append(hostSlc, host)
	}
	return hostSlc, nil
}

// LoadHostlist loads all entries from the file at the given path (designed for txt files) and
// returns them in the form of a slice of Host types. usrIsHost determines whether or not the ip is the
// same as username + ".local".
func LoadHostlist(path string, usrIsHost bool) ([]Host, error) {
	slc, err := fileToSlice(path)
	if err != nil {
		return nil, fmt.Errorf("internal/loadcfg: failed to open %s: %s", path, err.Error())
	}

	hSlc, err := SlcToHosts(slc, usrIsHost)
	if err != nil {
		return nil, err
	}

	return hSlc, nil
}

// ExportToFile exports a string to a file. One will be created if necessary.
func ExportToFile(str, path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		os.Create(path)
	}

	f, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	if _, err := f.WriteAt([]byte(str), info.Size()); err != nil {
		return err
	}
	return nil
}
