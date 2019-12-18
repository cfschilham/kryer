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
	jsonPath string
}

// Dict is used to load a dictionary of passwords from the configured dictionary file
type Dict struct {
	pwds []string
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

// JSONPath returns the value of jsonPath in a Config type
func (c *Config) JSONPath() string {
	return c.jsonPath
}

// Pwds returns the value of pwds in a Dict type
func (d *Dict) Pwds() []string {
	return d.pwds
}

// LoadConfig returns a config type based on the values in cfg/config.yml
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("../cfg")
	if err := viper.ReadInConfig(); err != nil {
		return &Config{}, errors.New("internal/loadcfg: failed to load config: " + err.Error())
	}

	c := &Config{
		verbose:  viper.GetBool("verbose"),
		mode:     viper.GetString("mode"),
		dictPath: viper.GetString("dict_path"),
		jsonPath: viper.GetString("json_path"),
	}
	return c, nil
}

// LoadDict loads all entries from dict.txt and returns them in the form of an array of strings
func LoadDict(path string) (*Dict, error) {
	f, err := os.Open("../" + path)
	if err != nil {
		return &Dict{}, errors.New("internal/loadcfg: failed to open dictionary: " + err.Error())
	}

	d := &Dict{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		d.pwds = append(d.pwds, sc.Text())
	}

	return d, nil
}
