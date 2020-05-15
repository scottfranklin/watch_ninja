package main

import (
    "flag"
    "bytes"
    "fmt"
    "io"
    "log"
    "os"
    "os/exec"
)

var build_watcher *Watcher
var file_watcher *Watcher

func main() {
    cd := flag.String("C", ".", "change to this directory before invoking ninja")
    flag.Parse()

    err := os.Chdir(*cd)
    if err != nil {
        panic(err)
    }

    build_watcher, err := NewWatcher()
    if err != nil {
        panic(err)
    }
    err = build_watcher.UpdateWatchList([][]byte{[]byte("build.ninja")});
    if err != nil {
        panic(err)
    }

    go build_watcher.Handle(Update)

    file_watcher, err = NewWatcher()
    if err != nil {
        panic(err)
    }

    Update("")

    file_watcher.Handle(Build)
}

func Update(_ string) {
    cmd := exec.Command("ninja", "-t", "targets", "rule")
    out, err := cmd.CombinedOutput()
    if err != nil {
        log.Println("Generate watchlist:", err)
        return
    }
    files := bytes.Split(out, []byte("\n"))
    file_watcher.UpdateWatchList(files)
}

func Build(f string) {
    fmt.Println("")
    cmd := exec.Command("ninja", f + "^")
    out, err := cmd.StdoutPipe()
    if err != nil {
        log.Println(err)
        return
    }
    cmd.Start()
    go io.Copy(os.Stdout, out)
    if cmd.Wait() == nil {
        fmt.Println("ok")
    }
}
