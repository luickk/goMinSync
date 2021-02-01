package syncClient

import (
	"time"
	"strconv"
	"os"
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
	ChangeStream chan util.FileChange
	fileWatcherSyncFreq time.Duration
}

func New() syncClient {
	var sC syncClient
	sC.fileWatcherSyncFreq = time.Second * 5
	sC.ChangeStream = make(chan util.FileChange)
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
				encodedChg, err = util.EncodeMsg(&chg)
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

func (sC *syncClient)StartSyncFromRemote(dir string) {
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
					if _, err := os.Stat(dir+changeMsg.RelPath); err == nil {
						err := os.Remove(dir+changeMsg.RelPath)
						if err != nil {
						    fmt.Println(err)
						}
						fmt.Println("sub receive: " + changeMsg.Ctype + ": " + dirPath)
					}
				case "added":
					if _, err := os.Stat(dir+changeMsg.RelPath); os.IsNotExist(err) {
						if changeMsg.IsDir {
							os.MkdirAll(dir+changeMsg.RelPath, os.ModePerm)
						} else {
							if sC.tlsEnabled {
								util.DownloadFile(dir+changeMsg.RelPath, "https://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/files/"+changeMsg.FileHash+"?token="+sC.token)
							} else {
								util.DownloadFile(dir+changeMsg.RelPath, "http://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/"+changeMsg.FileHash+"?token=empty")
							}
							fmt.Println("sub receive: " + changeMsg.Ctype + ": " + dirPath)
						}
					} else {
						fmt.Println("exists already")
					}
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

      changes, err := util.FindPathHashMapChange(pathHashMap, oldPathHashMap, dir)
			if err != nil {
				fmt.Println(err)
			}
      oldPathHashMap = pathHashMap
      for _, change := range changes {
        sc.ChangeStream <- change
      }
      fmt.Println("-----")
      time.Sleep(sc.fileWatcherSyncFreq)
    }
  }()
}
