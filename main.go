package main

import (
	"context"
	"log"
	"os"
	"os/signal"
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

	watcher.Add(app.dir)

	for {
		select{
		case <- ctx.Done():
			log.Println("Shutdown...")
			return nil

		case event := <- watcher.Events:
			if event.Has(fsnotify.Write | fsnotify.Remove | fsnotify.Create) {
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