package internal

import (
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type ProjectSyncConfig struct {
	Local  string   `json:"source" yaml:"source"`
	Remote string   `json:"dest" yaml:"dest"`
	Ignore []string `json:"ignore" yaml:"ignore"`
	Delay  int      `json:"delay_time" yaml:"delay_time"`
	SupportRemove bool `json:"support_remove" yaml:"support_remove"`

	// scp
	Protocol       string `json:"protocol" yaml:"protocol"`
	Host           string `json:"host" yaml:"host"`
	Port           int    `json:"port" yaml:"port"`
	UserName       string `json:"username" yaml:"username"`
	Password       string `json:"password" yaml:"passowrd"`
	PrivateKeyPath string `json:"private_key" yaml:"private_key"`
}

type Sync struct {
	config *ProjectSyncConfig
	sftp   *sftp.Client
	ssh    *ssh.Client
}

type Project struct {
	config   ProjectSyncConfig
	wg       *sync.WaitGroup
	watcher  *fsnotify.Watcher
	chRemove chan string
	chUpdate chan string
}
