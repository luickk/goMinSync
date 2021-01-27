package syncClient

import (
	"fmt"
	"time"

	"goMinSync/pkg/remoteCacheToGo/cacheClient"
  "goMinSync/pkg/util"
)

type syncClient struct {
	address string
	syncPort int
	fileServerPort int
	tlsEnabled bool

	cacheClient cacheClient.RemoteCache
	ChangeStream chan *util.FileChange
	fileWatcherSyncFreq time.Duration
}

func New() syncClient {
	var sC syncClient
	sC.fileWatcherSyncFreq = time.Second * 5
	sC.ChangeStream = make(chan *util.FileChange)
	return sC
}

// if token and cert are empty cache & file server will not be encrypted nor token protected
func (sC *syncClient)ConnectToRemoteInstance(address string, syncPort int, fileServerPort int, token string, rootCert string) error {
	// unencryptd (for testing purposes)
	client := cacheClient.New()

	if err := client.ConnectToCache(address, syncPort, token, rootCert); err != nil {
		return err
	}

  sC.cacheClient = client
	sC.address = address
	sC.syncPort = syncPort
	sC.fileServerPort = fileServerPort

	if token == "" && rootCert == "" {
		sC.tlsEnabled = false
	} else {
		sC.tlsEnabled = true
	}
	return nil
}

func (sC *syncClient)StartSyncToRemote() {
	var (
		err error
		encodedChg []byte
	)
	go func() {
		for {
			select {
			case chg := <- sC.ChangeStream:
				encodedChg, err = util.EncodeMsg(chg)
				if err != nil {
					return
				}
				fmt.Println(chg.Ctype + ":" + chg.DirPath)
				// writing to connected cache to key Dir-Path with value change-type
				sC.cacheClient.AddValByKey(chg.DirPath, encodedChg)

				if sC.tlsEnabled {

				} else {
					// util.PostUploadFile("http://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/upload", chg.DirPath, "dataLoL")
				}
			}
		}
	}()
}

func (sC *syncClient)StartSyncFromRemote() {
	var (
		dirPath string
	)
	changeMsg := new(util.FileChange)
	subCh := sC.cacheClient.Subscribe()
	go func() {
		for {
			select {
			case changed := <- subCh:
				if err := util.DecodeMsg(changeMsg, changed.Data); err != nil {
					return
				}
				dirPath = changed.Key
				switch changeMsg.Ctype {
				case "changed":
					fmt.Println("changed smth: " + changeMsg.FileHash + ": " + dirPath)
				case "removed":
					fmt.Println("changed smth: " + changeMsg.FileHash + ": " + dirPath)
				case "added":
					fmt.Println("changed smth: " + changeMsg.FileHash + ": " + dirPath)
				}
			}
		}
	}()
}

func (sc *syncClient)AddDir(dir string) {
  go func() {
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
        chg := new(util.FileChange)
        chg.Ctype = cType
			  chg.DirPath = path
				chg.FileHash, err = util.HashFile(path)
				if err != nil {
					return
				}
        chg.Timestamp = time.Now().String()
        sc.ChangeStream <- chg
      }
      fmt.Println("-----")
      time.Sleep(sc.fileWatcherSyncFreq)
    }
  }()
}
