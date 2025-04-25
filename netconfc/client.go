package netconfc

import (
	"bufio"
	"fmt"
	"os"

	"github.com/Juniper/go-netconf/netconf"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// Client encapsulates SSH/NETCONF data and methods
type Client struct {
	host string
	user string
	pass string
	conf *ssh.ClientConfig
	Sess *netconf.Session
}

// NewClient creates a new NETCONF client
// Return Client or error
func NewClient(host, user, pass string) (*Client, error) {
	c := &Client{
		host: host,
		user: user,
		pass: pass,
		conf: &ssh.ClientConfig{},
		Sess: &netconf.Session{},
	}

	if err := c.buildConfig(); err != nil {
		return nil, err
	}

	return c, nil
}

// Open opens a netconf session
// Returns error is session fails to open
func (c *Client) Open() error {
	s, err := netconf.DialSSH(c.host, c.conf)
	if err != nil {
		return fmt.Errorf("unable to dial host - %v", err)
	}

	c.Sess = s

	return nil
}

// Close closes the netconf session gracefully
// Returns error if session is not gracefully closed
func (c *Client) Close() error {
	if err := c.Sess.Close(); err != nil {
		return fmt.Errorf("failed to gracefully close session - %v", err)
	}

	return nil
}

// Execute exectues the provided RPC method
// Returns RPC reply or error
func (c *Client) Execute(m string) (*netconf.RPCReply, error) {
	r, err := c.Sess.Exec(netconf.RawMethod(m))
	if err != nil {
		return nil, err
	}

	return r, nil
}

// buildConfig builds a new ssh/netconf client config
// TODO -- add pubkey and agent support
func (c *Client) buildConfig() error {
	if c.user == "" {
		if err := c.getUser(); err != nil {
			return fmt.Errorf("failed to read username - %v", err)
		}
	}

	if c.pass == "" {
		err := c.getPass()
		if err != nil {
			return fmt.Errorf("failed to read password - %v", err)
		}
	}

	c.conf = netconf.SSHConfigPassword(c.user, c.pass)

	return nil
}

// getPass asks the user for a password
func (c *Client) getPass() error {
	fmt.Print("Password: ")
	bytePass, err := term.ReadPassword(0)
	if err != nil {
		return err
	}
	fmt.Println()

	c.pass = string(bytePass)

	return nil
}

// getUser asks the user for a username
func (c *Client) getUser() error {
	fmt.Print("Username: ")
	b := bufio.NewScanner(os.Stdin)
	b.Scan()
	if err := b.Err(); err != nil {
		return err
	}

	if b.Text() == "" {
		fmt.Println("blank username not allowed")
		c.getUser()
	}

	c.user = b.Text()

	return nil
}
