package sshsession

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/palantir/stacktrace"
	"golang.org/x/crypto/ssh"
)

const (
	DefaultSSHUser = "docker"
	DefaultSSHPort = "22"
)

func New(hostname string, keys []string) (*ssh.Session, error) {
	hostnameParts := strings.Split(hostname, "@")
	host := hostname
	user := DefaultSSHUser
	if len(hostnameParts) > 2 {
		return nil, stacktrace.NewError("hostname is of the wrong format")
	} else if len(hostnameParts) == 2 {
		user = hostnameParts[0]
		host = hostnameParts[1]
	}

	hostParts := strings.Split(host, ":")
	port := DefaultSSHPort
	if len(hostParts) > 2 {
		return nil, stacktrace.NewError("hostname is of the wrong format")
	} else if len(hostParts) == 2 {
		host = hostParts[0]
		port = hostParts[1]
	}

	keyring, err := MakeKeyring(keys)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create ssh keyring")
	}

	cfg := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{keyring},
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, port), cfg)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create ssh connection to '%s'", hostname)
	}

	sess, err := conn.NewSession()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create ssh session to '%s'", hostname)
	}

	return sess, nil
}

func MakeSigner(keyPath string) (ssh.Signer, error) {
	fp, err := os.Open(keyPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to open key file '%s'", keyPath)
	}
	defer fp.Close()

	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to read key file '%s'", keyPath)
	}

	signer, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to parse private key '%s'", keyPath)
	}

	return signer, nil
}

func MakeKeyring(keys []string) (ssh.AuthMethod, error) {
	signers := []ssh.Signer{}
	for _, key := range keys {
		signer, err := MakeSigner(key)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to make signer of key '%s'", key)
		}
		signers = append(signers, signer)
	}

	return ssh.PublicKeys(signers...), nil
}
