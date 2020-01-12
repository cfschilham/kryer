package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/cfschilham/autossh/internal/loadcfg"
	"github.com/cfschilham/autossh/internal/sshatk"
)

func main() {
	fmt.Println("Fork of: AutoSSH v1.1.3 - https://github.com/cfschilham/autossh")

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
			fmt.Print("Enter host (type 'exit' to exit): ")
			sc := bufio.NewScanner(os.Stdin)
			sc.Scan()
			input := sc.Text()
			if sc.Err() != nil {
				fmt.Fprintln(os.Stderr, sc.Err())
				continue
			}

			if input == "exit" {
				os.Exit(0)
			}

			// Use user input to create a new Host struct
			host, err := loadcfg.StrToHost(input, config.UsrIsHost())
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				continue
			}

			// The IP is resolved first so that it doesn't have to be resolved again for every single password in the
			// dictionary
			ip, err := host.ResolveIP()
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				continue
			}

			fmt.Printf("Attempting to connect to '%s@%s'...\n", host.Username(), host.IP())
			pwd, err := sshatk.SSHDict(ip, host.Username(), config.Port(), dict.Pwds(), config)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				continue
			}
			fmt.Printf("Password of '%s' found: %s\n", host.Username()+"@"+host.IP(), pwd)

			if config.OutputPath() != "" {
				s := fmt.Sprintf("%s@%s:%s", host.Username(), host.IP(), pwd)
				if err := loadcfg.ExportToFile(s, config.OutputPath()); err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
				}
			}
		}

	case "hostlist":
		fmt.Printf("Loading %s...\n", config.HostlistPath())
		hosts, err := loadcfg.LoadHostlist(config.HostlistPath(), config.UsrIsHost())
		if err != nil {
			log.Fatalln(err.Error())
		}

		// Loop through host list and append found username, hostname and password-combinations to a map
		var foundCredentials = map[string]string{}
		for i, host := range hosts {
			fmt.Printf("%d%% done\n", int(math.Round(float64(i)/float64(len(hosts))*100)))
			fmt.Printf("Attempting to connect to '%s@%s'...\n", host.Username(), host.IP())

			// The IP is resolved first so that it doesn't have to be resolved again for every single password in the
			// dictionary
			ip, err := host.ResolveIP()
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				continue
			}

			pwd, err := sshatk.SSHDict(ip, host.Username(), config.Port(), dict.Pwds(), config)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				continue
			}
			fmt.Printf("Password of '%s' found: '%s'\n", host.Username()+"@"+host.IP(), pwd)
			if config.OutputPath() != "" {
				s := fmt.Sprintf("%s@%s:%s", host.Username(), host.IP(), pwd)
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
				}
			}
			foundCredentials[host.IP()+"@"+host.Username()] = pwd
		}

		// Print all found combinations and export them to a file if that is selected in config.yml
		if len(foundCredentials) > 0 {
			fmt.Println("The following combinations were found: ")
			for host, pwd := range foundCredentials {
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
