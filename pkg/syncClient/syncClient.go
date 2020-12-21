package syncClient

import (
	"fmt"
	"time"

  "remoteCacheToGo/cmd/cacheClient"
  "goMinSync/pkg/util"
)

type fileChange struct {
  Ctype string
  DirPath string
  Timestamp string
}

type syncClient struct {
	cacheClient cacheClient.RemoteCache
	ChangeStream chan *fileChange
	fileWatcherSyncFreq time.Duration
}

func New(address string, port int) (syncClient, error) {
	var sC syncClient
	cClient, err := connectToRemoteCache(address, port)
	sC.fileWatcherSyncFreq = time.Second * 5
	sC.ChangeStream = make(chan *fileChange)
	if err != nil {
		fmt.Println(err)
		return sC, err
	}
	sC.cacheClient = cClient
	return sC, nil
}

func connectToRemoteCache(address string, port int) (cacheClient.RemoteCache, error) {
	client, err := cacheClient.New(address, port, false, "", "")
  if err != nil {
    fmt.Println(err)
    return client, err
  }
	return client, err
}


func (sc *syncClient)AddDir(dir string) {
  go func(){
    oldPathHashMap := make(map[string]string, 0)
    pathHashMap := make(map[string]string, 0)
    var err error
    for {
      pathHashMap, err = util.CreatePathHasMap(dir)
      if err != nil {
        fmt.Println(err)
      }

      changes := util.FindPathHashMapChange(pathHashMap, oldPathHashMap)
      oldPathHashMap = pathHashMap
      for path, cType := range changes {
        chg := new(fileChange)
        chg.Ctype = cType
        chg.DirPath = path
        chg.Timestamp = time.Now().String()
        sc.ChangeStream <- chg
      }
      fmt.Println("-----")
      time.Sleep(sc.fileWatcherSyncFreq)
    }
  }()
}
