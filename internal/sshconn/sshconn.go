package sshconn

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

const authErrSubstring = "ssh: handshake failed: ssh: unable to authenticate"

// SSHConn attempts to connect to the remote host using the passed values. The variable
// host should be an IP, username a remote username, port the remote port (typically 22)
// and pwd the corresponding password. A successful connection means a nil error is returned.
func SSHConn(host, username, port, pwd string) error {
	clientConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(pwd),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Attempt to connect to remote host.
	conn, err := ssh.Dial("tcp", host+":"+port, clientConfig)
	if err != nil {
		return fmt.Errorf("internal/sshconn: %s", err.Error())
	}
	conn.Close()
	return nil
}
