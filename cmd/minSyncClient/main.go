package main

import (
  "os"
  "path/filepath"
  "time"
  "fmt"

	"goMinSync/pkg/syncClient"
  "goMinSync/pkg/util"
)

const (
  added = iota
  removed = iota
  edited = iota
)

type fileChange struct {
  cType int
  fileHash string
  timestamp string
}

var fileWatcherSyncFreq time.Duration = time.Second * 1

func main() {
	_, err := syncClient.New("127.0.0.1", 8000)

  // should be located beneath err return of syncClient connect
  fileChangeWatcher("/Users/luickklippel/Documents/Temp")

  if err != nil {
    fmt.Println(err)
    return
  }

}

func fileChangeWatcher(dir string) chan *fileChange {
  fileChange := make(chan *fileChange)
   for {
      fileList := make([]string, 0)
      fileHashList := make([]string, 0)
      e := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
        fileList = append(fileList, path)
        return err
      })
      if e != nil {
        panic(e)
      }
      var hash string
      for _, file := range fileList {
        hash = util.Sha256DirObj(file)
        fileHashList = append(fileHashList, hash)
        fmt.Println(file + ": " + hash)
      }
      fmt.Println("-----")
      time.Sleep(fileWatcherSyncFreq)
    }
  return fileChange
}
