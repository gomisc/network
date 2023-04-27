package ssh

import (
	"net/url"
	"os"
	"path"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"gopkg.in/gomisc/colors.v1"
	"gopkg.in/gomisc/errors.v1"
)

const (
	defaultSessionTimeout = time.Minute * 10
	errNotConnected       = errors.Const("not connected to host")
)

type sshClient struct {
	signers         *signers
	hostkeyCallback ssh.HostKeyCallback
	colors          colors.Generator

	sync.RWMutex
	connections map[string]*ssh.Client
}

func New(keysFilter ...string) (Client, error) {
	cli := &sshClient{
		colors:      colors.NewGenerator(),
		connections: make(map[string]*ssh.Client),
	}

	sigs, err := newSigners(keysFilter...)
	if err != nil {
		return nil, errors.Wrap(err, "get signers")
	}

	cli.signers = sigs

	if cli.hostkeyCallback, err = knownhosts.New(path.Join(os.Getenv("HOME"), ".ssh", "known_hosts")); err != nil {
		return nil, errors.Wrap(err, "creating host key callback")
	}

	return cli, nil
}

func (cli *sshClient) Connect(endpoint, key string, o ...Option) error {
	var (
		conn *ssh.Client
		err  error
	)

	if conn = cli.connection(endpoint); conn == nil {
		if conn, err = cli.connect(endpoint, key, o...); err != nil {
			return errors.Wrap(err, "try connection to host")
		}
	}

	return nil
}

func (cli *sshClient) Session(endpoint string, o ...Option) (Session, error) {
	var conn *ssh.Client

	if conn = cli.connection(endpoint); conn == nil {
		return nil, errNotConnected
	}

	return cli.newSession(conn, o...)
}

func (cli *sshClient) connect(endpoint, key string, o ...Option) (*ssh.Client, error) {
	uri, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "parse endpoint")
	}

	opts := processOptions(o...)

	var signer ssh.Signer

	if signer = cli.signers.Signer(key); signer == nil {
		return nil, errors.Wrapf(err, "not found signer for key %s", key)
	}

	conf := ssh.ClientConfig{
		User: uri.User.Username(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.FixedHostKey(signer.PublicKey()),
		Timeout:         opts.sessionTimeOut,
	}

	var conn *ssh.Client

	if conn, err = ssh.Dial("tcp", uri.Host, &conf); err != nil {
		return nil, errors.Ctx().
			Str("host", uri.Host).
			Str("user", uri.User.Username()).
			Str("key", key).
			Wrap(err, "try host connection")
	}

	cli.Lock()
	defer cli.Unlock()

	cli.connections[endpoint] = conn

	return conn, nil
}

func (cli *sshClient) connection(endpoint string) *ssh.Client {
	cli.RLock()
	defer cli.RUnlock()

	if conn, ok := cli.connections[endpoint]; ok {
		return conn
	}

	return nil
}
