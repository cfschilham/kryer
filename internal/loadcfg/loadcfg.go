package loadcfg

import (
	"bufio"
	"errors"
	"os"

	"github.com/spf13/viper"
)

// Config is used to load config values from cfg/config.yml
type Config struct {
	verbose bool
	mode,
	dictPath,
	hostlistPath,
	jsonInPath,
	jsonOutPath string
}

// Dict is used to load a dictionary of passwords from the configured dictionary file
type Dict struct {
	pwds []string
}

type Hostlist struct {
	hosts []string
}

// Verbose returns the value of verbose in a Config type
func (c *Config) Verbose() bool {
	return c.verbose
}

// Mode returns the value of mode in a Config type
func (c *Config) Mode() string {
	return c.mode
}

// DictPath returns the value of dictPath in a Config type
func (c *Config) DictPath() string {
	return c.dictPath
}

// HostlistPath returns the value of hostlistPath in a Config type
func (c *Config) HostlistPath() string {
	return c.hostlistPath
}

// JSONInPath returns the value of jsonInPath in a Config type
func (c *Config) JSONInPath() string {
	return c.jsonInPath
}

// JSONOutPath returns the value of jsonOutPath in a Config type
func (c *Config) JSONOutPath() string {
	return c.jsonOutPath
}

// Pwds returns the value of pwds in a Dict type
func (d *Dict) Pwds() []string {
	return d.pwds
}

// Hosts returns the value of hosts in a Hostlist type
func (hl *Hostlist) Hosts() []string {
	return hl.hosts
}

// LoadConfig returns a config type based on the values in cfg/config.yml
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("../cfg")
	if err := viper.ReadInConfig(); err != nil {
		return &Config{}, errors.New("internal/loadcfg: failed to load config: " + err.Error())
	}

	c := &Config{
		verbose:      viper.GetBool("verbose"),
		mode:         viper.GetString("mode"),
		dictPath:     viper.GetString("dict_path"),
		hostlistPath: viper.GetString("hostlist_path"),
		jsonInPath:   viper.GetString("json_input_path"),
		jsonOutPath:  viper.GetString("json_output_path"),
	}
	return c, nil
}

func txtToStringArr(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return []string{}, err
	}

	sArr := []string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		sArr = append(sArr, sc.Text())
	}
	if sc.Err() != nil {
		return []string{}, sc.Err()
	}

	return sArr, nil
}

// LoadDict loads all entries from the file at the given path (designed for txt files) and returns them in the form of a Dictionary struct
func LoadDict(path string) (*Dict, error) {
	sArr, err := txtToStringArr(path)
	if err != nil {
		return &Dict{}, errors.New("internal/loadcfg: failed to open " + path + ": " + err.Error())
	}
	return &Dict{pwds: sArr}, nil
}

// LoadHostlist loads all entries from the file at the given path (designed for txt files) and returns them in the form of a Hostlist struct
func LoadHostlist(path string) (*Hostlist, error) {
	sArr, err := txtToStringArr(path)
	if err != nil {
		return &Hostlist{}, errors.New("internal/loadcfg: failed to open " + path + ": " + err.Error())
	}
	return &Hostlist{hosts: sArr}, nil
}
