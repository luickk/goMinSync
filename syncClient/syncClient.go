package syncClient

import (
	"time"
	"strconv"
	"fmt"

	"goMinSync/pkg/remoteCacheToGo/cacheClient"
  "goMinSync/pkg/util"
)

type syncClient struct {
	address string
	syncPort int
	token string
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

	sC.token = token
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
				fmt.Println(chg.Ctype + ":" + chg.RelPath)
				// writing to connected cache to key Dir-Path with value change-type
				sC.cacheClient.AddValByKey(chg.RelPath, encodedChg)
				if sC.tlsEnabled {
					if !chg.IsDir {
						_, err := util.PostUploadFile("http://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/upload?token="+sC.token+"&name="+chg.FileHash, chg.AbsPath, "file")
						if err != nil {
							return
						}
					}
				} else {
					if !chg.IsDir {
						_, err := util.PostUploadFile("http://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/upload?token=empty"+"&name="+chg.FileHash, chg.AbsPath, "file")
						if err != nil {
							return
						}
					}
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
					fmt.Println("sub receive: " + changeMsg.Ctype + ": " + dirPath)
				case "removed":
					fmt.Println("sub receive: " + changeMsg.Ctype + ": " + dirPath)
				case "added":
					fmt.Println("sub receive: " + changeMsg.Ctype + ": " + dirPath)
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
				return
      }

      changes := util.FindPathHashMapChange(pathHashMap, oldPathHashMap)
      oldPathHashMap = pathHashMap
      for path, cType := range changes {
        chg := new(util.FileChange)
        chg.Ctype = cType
			  chg.RelPath = path
			  chg.AbsPath = dir+chg.RelPath
				chg.FileHash, err = util.HashFile(chg.AbsPath)
				if err != nil {
					return
				}
				ok, err := util.IsDirectory(chg.AbsPath)
				if err != nil {
					return
				}
				chg.IsDir = ok
        chg.Timestamp = time.Now().String()
        sc.ChangeStream <- chg
      }
      fmt.Println("-----")
      time.Sleep(sc.fileWatcherSyncFreq)
    }
  }()
}
