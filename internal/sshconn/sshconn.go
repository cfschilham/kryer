package sshconn

import (
	"golang.org/x/crypto/ssh"
)

// SSHConn establishes an ssh connection and then executes the provided command in a shell
func SSHConn(host, user, pass, cmd string) error {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to remote port and open a new session
	conn, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return err
	}

	session, err := conn.NewSession()
	if err != nil {
		return err
	}

	// Open a remote shell in the session
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		session.Close()
		return err
	}

	// Execute cmd in remote shell
	if err := session.Run(cmd); err != nil {
		return err
	}

	session.Close()
	return nil
}
