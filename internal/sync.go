package internal

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func NewSyncConnect(config ProjectSyncConfig) (*sftp.Client, error) {
	var (
		sftpClient *sftp.Client
		auth       []ssh.AuthMethod
		err        error
	)

	sync := &Sync{
		config: &config,
	}

	if config.PrivateKeyPath == "" {
		auth = append(auth, ssh.Password(config.Password))
	} else {
		key, err := ioutil.ReadFile(config.PrivateKeyPath)
		if err != nil {
			return nil, errors.New("unable to read private key: " + err.Error())
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, errors.New("unable to parse private key: " + err.Error())
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	clientConfig := &ssh.ClientConfig{
		User:    config.UserName,
		Auth:    auth,
		Timeout: 30 * time.Second,
		//		HostKeyCallback: ssh.FixedHostKey(hostKey),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), clientConfig)
	if err != nil {
		return nil, err
	}
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}
	sync.sftp = sftpClient
	sync.ssh = sshClient
	log.Printf("[Info] connect %s:%d success", config.Host, config.Port)
	return sftpClient, nil
}
