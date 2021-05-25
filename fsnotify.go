package main

import (
	"flag"
	"log"
	"sync"

	"github.com/lpflpf/fsnotify/internal"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "", "")
	flag.Parse()

	configs := []internal.ProjectSyncConfig{}
	if err := internal.LoadConfig(configFile, &configs); err != nil {
		log.Fatal("load config failed: ", err.Error())
	}

	wg := &sync.WaitGroup{}

	for _, config := range configs {
		proj, err := internal.NewProject(config, wg)

		if err != nil {
			log.Fatalf("init Project %s failed", config.Local)
		}

		go proj.SyncRemove()
		go proj.SyncUpdate()

		go proj.Watch()
	}
	wg.Wait()
}
