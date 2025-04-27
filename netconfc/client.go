// netconfc/client.go

package netconfc

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Juniper/go-netconf/netconf"
	"golang.org/x/crypto/ssh"
)

// Client encapsulates SSH/NETCONF data and methods
type Client struct {
	host    string
	user    string
	pass    string
	port    uint16
	timeout time.Duration
	conf    *ssh.ClientConfig
	logging *log.Logger
	Sess    *netconf.Session
}

// Config will hold the configuration options for the NETCONF connection
type Config struct {
	Host    string
	User    string
	Pass    string
	Port    uint16
	Timeout time.Duration
	Logging *log.Logger
	// TODO add ssh agent and key configuration options
}

// NewClient checks the given config and creates a new client if valid
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("invalid netconfc config - config is nil")
	}
	if cfg.Host == "" {
		return nil, fmt.Errorf("invalid netconfc config - host is empty")
	}
	if cfg.User == "" {
		return nil, fmt.Errorf("invalid netconfc config - user is empty")
	}
	if cfg.Pass == "" {
		return nil, fmt.Errorf("invalid netconfc config - pass is empty")
	}
	if cfg.Timeout < 0 {
		return nil, fmt.Errorf("invalid netconfc config - timeout must be greater than 0")
	}
	if cfg.Port < 0 || cfg.Port == 0 {
		cfg.Port = 830
	}

	c := &Client{
		host:    cfg.Host,
		user:    cfg.User,
		pass:    cfg.Pass,
		port:    cfg.Port,
		timeout: cfg.Timeout,
		logging: cfg.Logging,
		conf:    &ssh.ClientConfig{},
		Sess:    &netconf.Session{},
	}

	if c.logging == nil {
		log.Println("no logger provided. logging to stdout")
		c.logging = log.New(os.Stdout, "", log.LstdFlags)
	}

	if c.timeout == 0 {
		c.timeout = 30 * time.Second
	}

	c.conf = netconf.SSHConfigPassword(c.user, c.pass)

	c.logging.Printf("created new netconf client:\n+ host: %s\n+ user: %s\n+ port: %d\n", c.host, c.user, c.port)

	return c, nil
}

// Open opens a netconf session
// Returns error is session fails to open
func (c *Client) Open() error {
	h := c.host
	if c.port != 830 {
		h = strings.Join([]string{c.host, fmt.Sprint(c.port)}, ":")
	}

	c.logging.Printf("dialing host %s\n", h)

	s, err := netconf.DialSSH(h, c.conf)
	if err != nil {
		return fmt.Errorf("unable to dial host - %w", err)
	}

	c.Sess = s
	return nil
}

// Close closes the netconf session gracefully
// Returns error if session is not gracefully closed
func (c *Client) Close() error {
	c.logging.Println("closing netconf session")
	if err := c.Sess.Close(); err != nil {
		return fmt.Errorf("failed to gracefully close session - %w", err)
	}

	return nil
}

// Execute exectues the provided RPC method
// Returns RPC reply or error
func (c *Client) Execute(m string) (*netconf.RPCReply, error) {
	c.logging.Printf("executing rpc method - %s", m)
	r, err := c.Sess.Exec(netconf.RawMethod(m))
	if err != nil {
		return nil, fmt.Errorf("failed to make rpc call %s - %w", m, err)
	}

	return r, nil
}

/* leave it to the program to ask the user for the user and password, not the library
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
*/
