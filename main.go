package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"

	"github.com/fsnotify/fsnotify"
)

type App struct {
	dir string
	uid int
	gid int
}

func (app App) Run(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = filepath.Walk(app.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				return err
			}
			log.Printf("Watching directory: %s", path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}


	for {
		select{
		case <- ctx.Done():
			log.Println("Shutdown...")
			return nil

		case event := <- watcher.Events:
			if event.Has(fsnotify.Write | fsnotify.Create) {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					log.Printf("New directory created: %s. Adding to watcher.", event.Name)
					err := watcher.Add(event.Name)
					if err != nil {
						log.Printf("Failed to add new directory: %v", err)
					}
				}

				log.Printf("%s Change Detected.\n", event.Name)
				if err := os.Chown(event.Name, app.uid, app.gid); err != nil {
					log.Println(err)
					continue
				}
				log.Printf("Executed %s Chown.\n", event.Name)
			}

		case err := <-watcher.Errors:
			if err != nil {
				return err
			}
		}
	}
}


func main() {
	args := os.Args[1:]
	if len(args) != 3 {
		Usage()
		os.Exit(0)
	}
	dir := args[0]
	if !Exists(dir) {
		log.Fatalln("dir does not exists.")
	}

	uid, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatalln("uid is int")
	}
	gid, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatalln("gid is int")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	app := App {
		dir: dir,
		uid: uid,
		gid: gid,
	}

	if err := app.Run(ctx); err != nil {
		log.Fatalln(err)
	}
}

func Usage() {
	log.Println("cmd [watch_dir] [uid] [gid]")
}

func Exists(filename string) bool {
    _, err := os.Stat(filename)
    return err == nil
}