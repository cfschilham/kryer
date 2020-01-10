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
	verbose,
	usrIsHost,
	multiThreaded,
	exportPwdToFile bool
	maxThreads int
	mode,
	port,
	dictPath,
	hostlistPath,
	pwdFilePath string
}

// Dict is used to load a dictionary of passwords from the configured dictionary file, Dict.pwds
// represents the data stored in the dict file.
type Dict struct {
	pwds []string
}

// Host is used to represent a combination of a username and an ip to connect to via SSH.
type Host struct {
	username,
	ip string
}

// Verbose returns the value of verbose in a Config type.
func (c Config) Verbose() bool {
	return c.verbose
}

// UsrIsHost returns the value of usrIsHost in a Config type.
func (c Config) UsrIsHost() bool {
	return c.usrIsHost
}

// MultiThreaded returns the value of multiThreaded in a Config type.
func (c Config) MultiThreaded() bool {
	return c.multiThreaded
}

// ExportPwdToFile returns the value of exportPwdToFile in a Config type
func (c Config) ExportPwdToFile() bool {
	return c.exportPwdToFile
}

// MaxThreads returns the value of maxThreads in a Config type.
func (c Config) MaxThreads() int {
	return c.maxThreads
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

// PwdFilePath returns the value of pwdFilePath in a Config type.
func (c Config) PwdFilePath() string {
	return c.pwdFilePath
}

// Pwds returns the value of pwds in a Dict type.
func (d *Dict) Pwds() []string {
	return d.pwds
}

// IP returns the value of ip in a Host type.
func (h Host) IP() string {
	return h.ip
}

// Username returns the value of username in a Host type.
func (h Host) Username() string {
	return h.username
}

func (h Host) ResolveIP() (string, error) {
	resolver := net.Resolver{
		PreferGo: false,
	}

	unparsedIPS, err := resolver.LookupHost(context.Background(), h.ip)
	var ips []net.IP
	for _, unparsedIP := range unparsedIPS {
		ips = append(ips, net.ParseIP(unparsedIP))
	}

	if err != nil {
		return "", err
	}

	for _, ip := range ips {
		if ip.To4() != nil {
			return ip.To4().String(), nil
		}
	}
	return "", fmt.Errorf("internal/loadcfg: failed to resolve host: '%s'", h.ip)
}

// LoadConfig returns a config type based on the values in cfg/config.yml.
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("cfg")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("internal/loadcfg: failed to load config: %s", err.Error())
	}

	c := &Config{
		verbose:         viper.GetBool("verbose"),
		usrIsHost:       viper.GetBool("user_is_host"),
		multiThreaded:   viper.GetBool("multi_threaded"),
		exportPwdToFile: viper.GetBool("export_pwd_to_file"),
		maxThreads:      viper.GetInt("max_threads"),
		mode:            viper.GetString("mode"),
		port:            viper.GetString("port"),
		dictPath:        viper.GetString("dict_path"),
		hostlistPath:    viper.GetString("hostlist_path"),
		pwdFilePath:     viper.GetString("pwd_file_path"),
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
// of a Host struct with username: 'user1', ip: 'user1.local'.
func StrToHost(str string, usrIsHost bool) (Host, error) {
	if str == "" {
		return Host{}, errors.New("internal/loadcfg: empty string passed")
	}

	if usrIsHost {
		return Host{
			username: str,
			ip:       str + ".local",
		}, nil
	}
	for i, char := range str {
		if string(char) == "@" && i != len(str)-1 {
			return Host{
				username: str[0:i],
				ip:       str[i+1:],
			}, nil
		}
	}
	return Host{}, fmt.Errorf("internal/loadcfg: invalid hostname '%s'", str)
}

// strSlcToHosts takes a slice of strings and returns a slice of hosts. These hostscan be used to append
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
