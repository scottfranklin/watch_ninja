package main

import (
    "bytes"
    "fmt"
    "io"
    "log"
    "os"
    "os/exec"
)

func main() {
    build_watcher, err := NewWatcher()
    if err != nil {
        panic(err)
    }
    err = build_watcher.UpdateWatchList([][]byte{[]byte("build.ninja")});
    if err != nil {
        panic(err)
    }

    watcher, err := NewWatcher()
    if err != nil {
        panic(err)
    }

    Update(watcher)

    for {
        select {
        case <-build_watcher.Modified:
            Update(watcher)
        case f := <-watcher.Modified:
            Build(f)
        }
    }
}

func Update(w *Watcher) {
    cmd := exec.Command("ninja", "-t", "targets", "rule")
    out, err := cmd.CombinedOutput()
    if err != nil {
        log.Println(err)
        return
    }
    files := bytes.Split(out, []byte("\n"))
    w.UpdateWatchList(files)
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
