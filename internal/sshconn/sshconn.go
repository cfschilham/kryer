package sshconn

import (
	"golang.org/x/crypto/ssh"
)

// SSHConn attempts to connect to a remote SSH port given the host ip, port, username
// and password. A returned error of nil indicates a successful connection.
func SSHConn(host, user, port, pass string) error {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to remote port and open a new session
	conn, err := ssh.Dial("tcp", host+":"+port, sshConfig)
	if err != nil {
		return err
	}

	session, err := conn.NewSession()
	if err != nil {
		return err
	}

	session.Close()
	return nil
}
