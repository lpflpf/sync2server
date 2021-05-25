package internal

import (
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type ProjectSyncConfig struct {
	Local       string   `json:"source"`
	Remote      string   `json:"dest"`
	WatcherPath string   `json:"watcher_path"`
	Ignore      []string `json:"ignore"`
	AutoDelete  bool     `json:"auto_delete"`
	Delay       int      `json:"delay_time"`

	// scp
	Protocol       string `json:"protocol"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	UserName       string `json:"username"`
	Password       string `json:"password"`
	PrivateKeyPath string `json:"private_key"`
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
