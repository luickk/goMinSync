package main

import (
  "os"
  "path/filepath"
  "time"
  "fmt"

	"goMinSync/pkg/syncClient"
  "goMinSync/pkg/util"
)

type fileChange struct {
  Ctype string
  DirPath string
  Timestamp string
}

var fileWatcherSyncFreq time.Duration = time.Second * 5

func main() {
	_, err := syncClient.New("127.0.0.1", 8000)

  // should be located beneath err return of syncClient connect
  fileChange := fileChangeWatcher("/Users/luickklippel/Documents/Temp")
  for {
    select {
    case chg := <-fileChange:
      fmt.Println(chg.Timestamp + " " + chg.Ctype + ":" + chg.DirPath)
    }
  }
  if err != nil {
    fmt.Println(err)
    return
  }

}

func fileChangeWatcher(dir string) chan *fileChange {
  fileChangeCh := make(chan *fileChange)
  pathHashMap := make(map[string]string, 0)
  oldPathHashMap := make(map[string]string, 0)
  var hash string
  for {
    err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
      fi, err := os.Stat(path)
      if err != nil {
        fmt.Println(err)
      }
      if fi.Mode().IsRegular() {
        hash = util.Sha256DirObj(path)
        pathHashMap[path] = hash
      } else {
        pathHashMap[path] = ""
      }
      return err
    })
    if err != nil {
      fmt.Println(err)
    }
    changes := util.FindPathHashMapChange(pathHashMap, oldPathHashMap)
    for path, cType := range changes {
      chg := new(fileChange)
      chg.Ctype = cType
      chg.DirPath = path
      chg.Timestamp = time.Now().String()
      fileChangeCh <- chg
    }
    for change, path := range changes {
      fmt.Println(change + ": " + path)
    }

    fmt.Println("-----")
    oldPathHashMap = pathHashMap
    time.Sleep(fileWatcherSyncFreq)
  }
  return fileChangeCh
}
