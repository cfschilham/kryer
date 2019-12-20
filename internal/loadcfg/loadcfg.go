package loadcfg

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config is used to load config values from cfg/config.yml
type Config struct {
	verbose,
	usrIsHost bool
	mode,
	port,
	dictPath,
	hostlistPath string
}

// Dict is used to load a dictionary of passwords from the configured dictionary file,
// Dict.pwds represents the data stored in the dict file.
type Dict struct {
	pwds []string
}

// Host is used exclusively in a Hostlist type to represent a combination of a username
// and an ip to connect to via SSH.
type Host struct {
	username,
	ip string
}

// Hostlist is used to load a list of hostnames from the configured hostlist file,
// Hostlist.hosts represents the data stored in the hostlist file, with separated
// usernames and ip's.
type Hostlist struct {
	hosts []Host
}

// Verbose returns the value of verbose in a Config type.
func (c *Config) Verbose() bool {
	return c.verbose
}

// UsrIsHost returns the value of usrIsHost in a Config type.
func (c *Config) UsrIsHost() bool {
	return c.usrIsHost
}

// Mode returns the value of mode in a Config type.
func (c *Config) Mode() string {
	return c.mode
}

// Port returns the value of port in a Config type.
func (c *Config) Port() string {
	return c.port
}

// DictPath returns the value of dictPath in a Config type.
func (c *Config) DictPath() string {
	return c.dictPath
}

// HostlistPath returns the value of hostlistPath in a Config type.
func (c *Config) HostlistPath() string {
	return c.hostlistPath
}

// Pwds returns the value of pwds in a Dict type.
func (d *Dict) Pwds() []string {
	return d.pwds
}

// IP returns the value of ip in a Host type.
func (h *Host) IP() string {
	return h.ip
}

// Username returns the value of username in a Host type.
func (h *Host) Username() string {
	return h.username
}

// Hosts returns the value of hosts in a Hostlist type.
func (hl *Hostlist) Hosts() []Host {
	return hl.hosts
}

// LoadConfig returns a config type based on the values in cfg/config.yml.
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("cfg")
	if err := viper.ReadInConfig(); err != nil {
		return &Config{}, fmt.Errorf("internal/loadcfg: failed to load config: %s", err.Error())
	}

	c := &Config{
		verbose:      viper.GetBool("verbose"),
		usrIsHost:    viper.GetBool("user_is_host"),
		mode:         viper.GetString("mode"),
		port:         viper.GetString("port"),
		dictPath:     viper.GetString("dict_path"),
		hostlistPath: viper.GetString("hostlist_path"),
	}
	return c, nil
}

// fToStrSlc opens a file at the given path and returns a string slice.
// For every line of the file a new entry is appended to the string slice.
func fToStrSlc(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return []string{}, err
	}
	defer f.Close()

	slc := []string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		slc = append(slc, sc.Text())
	}
	if sc.Err() != nil {
		return []string{}, sc.Err()
	}

	return slc, nil
}

// LoadDict loads all entries from the file at the given path (designed for txt files) and returns
// them in the form of a Dictionary struct.
func LoadDict(path string) (*Dict, error) {
	slc, err := fToStrSlc(path)
	if err != nil {
		return &Dict{}, fmt.Errorf("internal/loadcfg: failed to open %s: %s", path, err.Error())
	}
	return &Dict{pwds: slc}, nil
}

// StrToHost takes a string and returns a host. Strings should be passed in the form 'username@host'.
func StrToHost(str string, usrIsHost bool) (Host, error) {
	if usrIsHost {
		return Host{
			username: str,
			ip:       str + ".local",
		}, nil
	}
	for i, char := range str {
		if string(char) == "@" {
			return Host{
				username: str[0:i],
				ip:       str[i+1:],
			}, nil
		}
	}
	return Host{}, fmt.Errorf("internal/loadcfg: invalid hostname '%s'", str)
}

// strSlcToHosts takes a slice of strings and returns a slice of hosts. These hosts
// can be used to append to a Hostlist type. Strings should be passed in the form 'username@host'.
func strSlcToHosts(slc []string, usrIsHost bool) ([]Host, error) {
	var hostSlc []Host
	for _, str := range slc {
		host, err := StrToHost(str, usrIsHost)
		if err != nil {
			return []Host{}, err
		}
		hostSlc = append(hostSlc, host)
	}
	return hostSlc, nil
}

// LoadHostlist loads all entries from the file at the given path (designed for txt files) and
// returns them in the form of a Hostlist struct. usrIsHost determines whether or not the ip is the
// same as username + ".local".
func LoadHostlist(path string, usrIsHost bool) (*Hostlist, error) {
	slc, err := fToStrSlc(path)
	if err != nil {
		return &Hostlist{}, fmt.Errorf("internal/loadcfg: failed to open %s: %s", path, err.Error())
	}

	hosts, err := strSlcToHosts(slc, usrIsHost)
	if err != nil {
		return nil, err
	}

	return &Hostlist{hosts: hosts}, nil
}
