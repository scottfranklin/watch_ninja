package main

import (
    "github.com/dersebi/golang_exp/exp/inotify"
    "log"
    "strings"
    "time"
)

type Watcher struct {
    watch *inotify.Watcher
    watch_list map[string]bool
}

func NewWatcher() (*Watcher, error) {
    watch, err := inotify.NewWatcher()
    if err != nil {
        return nil, err
    }
    w := &Watcher{watch, map[string]bool{}}
    return w, nil
}

func (w *Watcher) Handle(f func(string)) {
    last := ""
    next := time.Now()
    for {
        select {
        case e := <-w.watch.Error:
            log.Println("Inotify error:", e)
        case e := <-w.watch.Event:
            if e.Mask & inotify.IN_MODIFY != 0 {
                t := time.Now()
                if t.Before(next) && e.Name == last {
                    continue
                }
                next = t.Add(1 * time.Second)
                last = e.Name
                go f(e.Name)
            }
        }
    }
}

func (w *Watcher) UpdateWatchList(files [][]byte) error {
    watch_list := map[string]bool{}
    var err error

    for _, b := range files {
        f := strings.TrimSpace(string(b))
        if f != "" {
            if w.watch_list[f] {
                delete(w.watch_list, f)
                watch_list[f] = true
            } else {
                log.Println("Watching", f)
                if err = w.watch.AddWatch(f, inotify.IN_MODIFY); err != nil {
                    log.Println(err)
                } else {
                    watch_list[f] = true
                }
            }
        }
    }

    // unwatch leftovers
    for f := range w.watch_list {
        log.Println("Unwatching", f)
        if err = w.watch.RemoveWatch(f); err != nil {
            log.Println(err)
        }
    }

    w.watch_list = watch_list
    return err
}
