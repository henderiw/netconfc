package netconf

import (
	"io"

	"golang.org/x/crypto/ssh"
)

type sshClient struct {
	client *ssh.Client
}

// sshSessionTransport defines an SSH-based NETCONF session
type sshSessionTransport struct {
	sshSession *ssh.Session
	reader     io.Reader
	writer     io.WriteCloser
}

// NewClientSSH creates a new NETCONF SSH client from a
func NewClientSSH(client *ssh.Client) Client {
	return &sshClient{client: client}
}

func dialSSH(addr string, username string, authMethod ssh.AuthMethod, cb ssh.HostKeyCallback) (Client, error) {
	config := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: cb,
	}
	client, err := ssh.Dial("tcp", addr, config)

	if err != nil {
		return nil, err
	}

	return NewClientSSH(client), nil
}

// DialSSHWithPassword is a convenience function to creating a new NETCONF over SSH session
func DialSSHWithPassword(addr string, username string, password string, cb ssh.HostKeyCallback) (Client, error) {
	return dialSSH(addr, username, ssh.Password(password), cb)
}

// DialSSHWithPublicKey is a convenience function to creating a new NETCONF over SSH session
func DialSSHWithPublicKey(addr string, username string, signer ssh.Signer, cb ssh.HostKeyCallback) (Client, error) {
	return dialSSH(addr, username, ssh.PublicKeys(signer), cb)
}

// NewSession creates a new session from the given client
func (c *sshClient) NewSession() (*Session, error) {
	var s sshSessionTransport
	var session *Session
	var err error

	if s.sshSession, err = c.client.NewSession(); err == nil {
		if session, err = s.init(); err == nil {
			return session, nil
		}
		s.sshSession.Close()
	}

	return nil, err
}

// Close SSH session
func (c *sshClient) Close() error {
	return c.client.Close()
}

func (s *sshSessionTransport) init() (*Session, error) {
	var err error

	if s.writer, err = s.sshSession.StdinPipe(); err != nil {
		return nil, err
	}

	if s.reader, err = s.sshSession.StdoutPipe(); err != nil {
		return nil, err
	}

	if err = s.sshSession.RequestSubsystem("netconf"); err != nil {
		return nil, err
	}

	return newSession(s)
}

func (s sshSessionTransport) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}

func (s sshSessionTransport) Write(p []byte) (n int, err error) {
	return s.writer.Write(p)
}

func (s sshSessionTransport) Close() error {
	s.writer.Close()
	return s.sshSession.Close()
}