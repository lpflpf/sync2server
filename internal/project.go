package internal

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/sftp"
)

func NewProject(config ProjectSyncConfig, wg *sync.WaitGroup) (*Project, error) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ignores := []string{}
	for _, ignore := range config.Ignore {
		abs, _ := filepath.Abs(filepath.Join(config.Local, ignore))
		ignores = append(ignores, abs)
	}
	config.Ignore = ignores

	wg.Add(3)
	return &Project{
		config:   config,
		watcher:  watcher,
		chRemove: make(chan string, 1000),
		chUpdate: make(chan string, 1000),
		wg:       wg,
	}, nil
}

func (proj *Project) Watch() {
	proj.wg.Done()

	filepath.Walk(proj.config.Local, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			path, err := filepath.Abs(path)
			if Ignore(proj.config, path) {
				return nil
			}
			if err != nil {
				log.Printf("add watch failed, path: %s, err: %v", path, err)
			}
			err = proj.watcher.Add(path)
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	})
	for {
		select {
		case event, ok := <-proj.watcher.Events:
			if !ok {
				return
			}

			log.Println("[monior]", event)

			if Ignore(proj.config, event.Name) {
				continue
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				proj.chUpdate <- event.Name
				continue
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				info, _ := os.Stat(event.Name)
				if info.IsDir(){
					continue
				}


				proj.chUpdate <- event.Name
			}

			if event.Op&fsnotify.Remove == fsnotify.Remove {
				proj.chRemove <- event.Name
			}

			if event.Op&fsnotify.Rename == fsnotify.Rename {
				proj.chRemove <- event.Name
			}

		case err, ok := <-proj.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("project: %s, failed:%s", proj.config.Local, err)
		}
	}
}

type None struct{}

func (proj *Project) SyncRemove() {
	defer proj.wg.Done()
	for filename := range proj.chRemove {
		log.Printf("[info] remove file : %s", filename)
	}

}

func (proj *Project) SyncUpdate() {
	defer proj.wg.Done()

	filenames := make(map[string]None)
	duration := time.Second * time.Duration(proj.config.Delay)
	ticker := time.NewTimer(duration)
	ticker.Stop()
	for filename := range proj.chUpdate {
		filenames[filename] = None{}
		ticker.Reset(duration)

		// 收集文件，批量同步
	stopRecv:
		for {
			select {
			case <-ticker.C:
				break stopRecv
			case filename := <-proj.chUpdate:
				filenames[filename] = None{}
			}
		}

		sftpClient, err := NewSyncConnect(proj.config)
		if err != nil {
			continue
		}

		for filename := range filenames {
			log.Println("begin upload ", filename)
			proj.Upload(sftpClient, filename)
		}

		sftpClient.Close()

		// 清空数据
		filenames = make(map[string]None)
	}
}

func Ignore(config ProjectSyncConfig, file string) bool {
	file, _ = filepath.Abs(file)
	now := time.Now()
	for _, ignore := range config.Ignore {
		if len(ignore) <= len(file) {
			if file[0:len(ignore)] == ignore {
				return true
			}
		}

		stat, err := os.Stat(file)
		if err != nil || (!stat.IsDir() && now.Sub(stat.ModTime()) > 3*time.Minute) {
			return true
		}
	}
	return false
}

func (proj *Project) Upload(sftpClient *sftp.Client, localFilePath string) {
	if f, err := os.Stat(localFilePath); err == nil {
		if f.IsDir() {
			proj.UploadDirectory(sftpClient, localFilePath)
		} else {
			proj.UploadFile(sftpClient, localFilePath)
		}
	}
}

// UploadFile
func (proj *Project) UploadFile(sftpClient *sftp.Client, localFilePath string) {
	if Ignore(proj.config, localFilePath) {
		return
	}
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer srcFile.Close()

	remoteFilePath := path.Join(proj.config.Remote, localFilePath[len(proj.config.Local):])
	remoteFilePath = strings.ReplaceAll(remoteFilePath, "\\", "/")
	dstFile, err := sftpClient.Create(remoteFilePath)
	if err != nil {
		log.Printf("[Error] mkdir file %s failed", localFilePath)
		return
	}
	defer dstFile.Close()

	buf, err := ioutil.ReadAll(srcFile)

	if err != nil {
		panic(err)
	}

	dstFile.Write(buf)
	log.Printf("[Info] upload file: src: %s, remote: %s:%s", localFilePath, proj.config.Host, remoteFilePath)
}

// UploadDirectory
func (proj *Project) UploadDirectory(sftpClient *sftp.Client, localPath string) {
	if Ignore(proj.config, localPath) {
		return
	}

	proj.watcher.Add(localPath)
	localFiles, err := ioutil.ReadDir(localPath)

	if err != nil {
		panic(err)
	}

	changedFileName := localPath[len(proj.config.Local):]

	sftpClient.Mkdir(path.Join(proj.config.Remote, changedFileName))

	log.Print("[Info] make remote dir: ", path.Join(proj.config.Remote, changedFileName))

	for _, backupDir := range localFiles {
		localFilePath := path.Join(localPath, backupDir.Name())
		remoteFilePath := path.Join(proj.config.Remote, changedFileName, backupDir.Name())
		remoteFilePath = strings.ReplaceAll(remoteFilePath, "\\", "/")

		if backupDir.IsDir() {
			if Ignore(proj.config, localFilePath) {
				continue
			}
			proj.watcher.Add(localFilePath)
			if err := sftpClient.Mkdir(remoteFilePath); err != nil {
				log.Printf("sync file %s failed, err:%s", remoteFilePath, err)
			}
			log.Printf("Mkdir directory %s:%s", proj.config.Host, remoteFilePath)
			proj.UploadDirectory(sftpClient, localFilePath)
		} else {
			proj.UploadFile(sftpClient, path.Join(localPath, backupDir.Name()))
		}
	}
}
