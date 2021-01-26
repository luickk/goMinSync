package syncClient

import (
	"fmt"
	"time"
	"log"
	"goMinSync/pkg/remoteCacheToGo/cacheClient"
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

func New() syncClient {
	var sC syncClient
	sC.fileWatcherSyncFreq = time.Second * 5
	sC.ChangeStream = make(chan *fileChange)
	return sC
}

func (sC *syncClient)ConnectToRemoteInstance(address string, port int) error {
	// unencryptd (for testing purposes)
	client, err := cacheClient.New(address, port, false, "", "")
  if err != nil {
    return err
  }
	sC.cacheClient = client
	return nil
}

func (sC *syncClient)StartSyncToRemoteInstance() {
	go func() {
		var (
			dirPath string
			changeType string
		)
		subCh := sC.cacheClient.Subscribe()
		for {
			select {
			case chg := <- sC.ChangeStream:
				fmt.Println(chg.Ctype + ":" + chg.DirPath)
				// writing to connected cache to key Dir-Path with value change-type
				sC.cacheClient.AddValByKey(chg.DirPath, []byte(chg.Ctype))
			case changed := <- subCh:
				changeType = string(changed.Data)
				dirPath = changed.Key
				switch changeType {
				case "changed":
					fmt.Println("changed smth: " + changeType + ": " + dirPath)
				case "removed":
					fmt.Println("changed smth: " + changeType + ": " + dirPath)
				case "added":
					fmt.Println("changed smth: " + changeType + ": " + dirPath)
				}
			}
		}
	}()
}

func (sc *syncClient)AddDir(dir string) {
  go func(){
    oldPathHashMap := make(map[string]string, 0)
    pathHashMap := make(map[string]string, 0)
    var err error
    for {
      pathHashMap, err = util.CreatePathHashMap(dir)
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
