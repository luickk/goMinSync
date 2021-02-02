package syncClient

import (
	"time"
	"strconv"
	"os"
  "math/rand"
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
	sC.fileWatcherSyncFreq = time.Second * 1
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

	if err := sC.StartSyncToRemote(); err != nil {
		return err
	}

	return nil
}

func (sC *syncClient)StartSyncToRemote() error {
	var (
		err error
		encodedChg []byte
		url string
	)
	go func() error {
		for {
			select {
			case chg := <- sC.ChangeStream:
				encodedChg, err = util.EncodeMsg(&chg)
				if err != nil {
					return err
				}
				url = ""
				if sC.tlsEnabled {
					url = "http://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/upload?token="+sC.token+"&name="+chg.FileHash
				} else {
					url = "http://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/upload?token=empty"+"&name="+chg.FileHash
				}
				_, err = os.Stat(chg.AbsPath)
				if !chg.IsDir && err == nil {
					_, err := util.PostUploadFile(url, chg.AbsPath, "file")
					if err != nil {
						return err
					}
				}
				// writing to connected cache to key Dir-Path with value change-type
				sC.cacheClient.AddValByKey(chg.RelPath, encodedChg)
			}
		}
	}()
	return nil
}

func (sC *syncClient)StartSyncFromRemote(dir string, syncGroup int, clientOrigin int) error {
	var (
		dirPath string
		url string
	)
	changeMsg := new(util.FileChange)
	subCh := sC.cacheClient.Subscribe()
	go func() error {
		for {
			select {
			case dbSub := <- subCh:
				if err := util.DecodeMsg(changeMsg, dbSub.Data); err != nil {
					return err
				}
				if changeMsg.SyncGroup == syncGroup && changeMsg.OriginId != clientOrigin {
					fmt.Println(changeMsg.OriginId)
					dirPath = dbSub.Key
					switch changeMsg.Ctype {
					case "changed":
						if _, err := os.Stat(dir+changeMsg.RelPath); err == nil {
							err := os.Remove(dir+changeMsg.RelPath)
							if err != nil {
							    return err
							}
						}
						url = ""
						if _, err := os.Stat(dir+changeMsg.RelPath); os.IsNotExist(err) {
							if changeMsg.IsDir {
								os.MkdirAll(dir+changeMsg.RelPath, os.ModePerm)
							} else {
								if sC.tlsEnabled {
									url = "https://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/files/"+changeMsg.FileHash+"?token="+sC.token
								} else {
									url = "http://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/files/"+changeMsg.FileHash+"?token=empty"
								}
								if err := util.DownloadFile(dir+changeMsg.RelPath, url); err != nil {
									return err
								}
							}
						}
					case "removed":
						if _, err := os.Stat(dir+changeMsg.RelPath); err == nil {
							err := os.Remove(dir+changeMsg.RelPath)
							if err != nil {
								return err
							}
						}
					case "added":
						url = ""
						if _, err := os.Stat(dir+changeMsg.RelPath); os.IsNotExist(err) {
							if changeMsg.IsDir {
								os.MkdirAll(dir+changeMsg.RelPath, os.ModePerm)
							} else {
								if sC.tlsEnabled {
									url = "https://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/files/"+changeMsg.FileHash+"?token="+sC.token
								} else {
									url = "http://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/files/"+changeMsg.FileHash+"?token=empty"
								}
								if err := util.DownloadFile(dir+changeMsg.RelPath, url); err != nil {
									return err
								}
							}
						}
					}
				}
			}
		}
	}()
	return nil
}

func (sc *syncClient)ChangeListener(dir string, syncGroup int, clientOrigin int) error {
  go func() error {
    oldPathHashMap := make(map[string]string, 0)
    pathHashMap := make(map[string]string, 0)
    var err error
    for {
      pathHashMap, err = util.CreatePathHashMap(dir)
      if err != nil {
				return err
      }
      changes, err := util.FindPathHashMapChange(pathHashMap, oldPathHashMap, dir)
			if err != nil {
				return err
			}
      oldPathHashMap = pathHashMap
      for _, change := range changes {
				change.SyncGroup = syncGroup
				change.OriginId = clientOrigin
        sc.ChangeStream <- change
      }
      time.Sleep(sc.fileWatcherSyncFreq)
    }
  }()
	return nil
}


func (sc *syncClient)AddDir(dir string, syncGroup int) error {
	clientOrigin := rand.Intn(1000)
	fmt.Println(dir + "-" + strconv.Itoa(clientOrigin))

	// "/Users/luickklippel/Documents/Temp Local"
	if err := sc.ChangeListener(dir, syncGroup, clientOrigin); err != nil {
		return err
	}
	if err := sc.StartSyncFromRemote(dir, syncGroup, clientOrigin); err != nil {
		return err
	}
	return nil
}
