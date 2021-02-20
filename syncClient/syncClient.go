package syncClient

import (
	"time"
	"strconv"
	"os"
  "math/rand"
  "sync"

	"goMinSync/pkg/remoteCacheToGo/cacheClient"
  "goMinSync/pkg/util"
)

type syncClient struct {
	address string
	syncPort int
	token string
	fileServerPort int
	tlsEnabled bool
	cert string

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
func (sC *syncClient)ConnectToRemoteInstance(address string, syncPort int, fileServerPort int, token string, cert string, errorStream chan error) {
	// unencryptd (for testing purposes)
	client := cacheClient.New()

	client.ConnectToCache(address, syncPort, token, cert, errorStream)
	sC.cert = cert
	sC.token = token
  sC.cacheClient = client
	sC.address = address
	sC.syncPort = syncPort
	sC.fileServerPort = fileServerPort

	if token == "" || cert == "" {
		sC.tlsEnabled = false
	} else {
		sC.tlsEnabled = true
	}

	sC.StartSyncToRemote(errorStream)

	return
}

func (sC *syncClient)StartSyncToRemote(errorStream chan error) {
	var (
		err error
		encodedChg []byte
		url string
	)
	go func() {
		for {
			select {
			case chg := <- sC.ChangeStream:
				encodedChg, err = util.EncodeMsg(&chg)
				if err != nil {
					errorStream <- err
					return
				}
				if sC.tlsEnabled {
					url = "https://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/upload?token="+sC.token+"&name="+chg.FileHash
				} else {
					url = "http://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/upload?token=empty"+"&name="+chg.FileHash
				}
				_, err = os.Stat(chg.AbsPath)
				if !chg.IsDir && err == nil {
					go func(absPath string, url string) {
						_, err := util.PostUploadFile(url, absPath, "file", sC.cert)
						if err != nil {
							errorStream <- err
							return
						}
						return
					}(chg.AbsPath, url)
				}
				// writing to connected cache to key Dir-Path with value change-type
				sC.cacheClient.AddValByKey(chg.RelPath, encodedChg)
			}
		}
	}()
	return
}

func (sC *syncClient)SyncFromRemote(dir string, pingPongSync chan util.FileChange, syncGroup int, clientOrigin int, errorStream chan error) {
	var (
		absPath string
		url string
	)
	changeMsg := new(util.FileChange)
	subCh := sC.cacheClient.Subscribe()
	for {
		select {
		case dbSub := <- subCh:
			if err := util.DecodeMsg(changeMsg, dbSub.Data); err != nil {
				errorStream <- err
				return
			}
			absPath = dir+changeMsg.RelPath
			if changeMsg.SyncGroup == syncGroup && changeMsg.OriginId != clientOrigin {
				pingPongSync <- *changeMsg
				switch changeMsg.Ctype {
				case "changed":
					if _, err := os.Stat(absPath); err == nil {
						err := os.Remove(absPath)
						if err != nil {
								errorStream <- err
						    return
						}
					}
					url = ""
					if _, err := os.Stat(absPath); os.IsNotExist(err) {
						if changeMsg.IsDir {
							os.MkdirAll(absPath, os.ModePerm)
						} else {
							if sC.tlsEnabled {
								url = "https://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/files/"+changeMsg.FileHash+"?token="+sC.token
							} else {
								url = "http://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/files/"+changeMsg.FileHash+"?token=empty"
							}
							go func(absPath string, url string) {
								if err := util.DownloadFile(absPath, url, sC.cert); err != nil {
									errorStream <- err
									return
								}
								return
							}(absPath, url)
						}
					}
				case "removed":
					if _, err := os.Stat(absPath); err == nil {
						err := os.Remove(absPath)
						if err != nil {
							errorStream <- err
							return
						}
					}
				case "added":
					url = ""
					if _, err := os.Stat(absPath); os.IsNotExist(err) {
						if changeMsg.IsDir {
							os.MkdirAll(absPath, os.ModePerm)
						} else {
							if sC.tlsEnabled {
								url = "https://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/files/"+changeMsg.FileHash+"?token="+sC.token
							} else {
								url = "http://"+sC.address+":"+strconv.Itoa(sC.fileServerPort)+"/files/"+changeMsg.FileHash+"?token=empty"
							}
							go func(absPath string, url string) {
								if err := util.DownloadFile(absPath, url, sC.cert); err != nil {
									errorStream <- err
									return
								}
								return
							}(absPath, url)
						}
					}
				}
			}
		}
	}
	return
}

func (sc *syncClient)ChangeListener(dir string, pingPongSync chan util.FileChange, syncGroup int, clientOrigin int, errorStream chan error) {
	pingPongSyncMap := make(map[string]*util.FileChange)
	pingPongSyncMapMutex := &sync.RWMutex{}

	go func() {
		for {
			select {
			case ppCh := <- pingPongSync:
				pingPongSyncMapMutex.Lock()
				pingPongSyncMap[dir+ppCh.RelPath] = &ppCh
				pingPongSyncMapMutex.Unlock()
			}
		}
	}()

  go func() {
    oldPathHashMap := make(map[string]string, 0)
    pathHashMap := make(map[string]string, 0)
		isPong := false
    var err error
    for {
      pathHashMap, err = util.CreatePathHashMap(dir)
      if err != nil {
				errorStream <- err
      }
      changes, err := util.FindPathHashMapChange(pathHashMap, oldPathHashMap, dir)
			if err != nil {
				errorStream <- err
			}
      oldPathHashMap = pathHashMap
      for _, change := range changes {

				isPong = false
				change.SyncGroup = syncGroup
				change.OriginId = clientOrigin

				pingPongSyncMapMutex.RLock()
				for _, ppChange := range pingPongSyncMap {
					if change.RelPath == ppChange.RelPath && change.FileHash == ppChange.FileHash && ppChange.Ctype == change.Ctype {
						isPong = true
					}
				}
				pingPongSyncMapMutex.RUnlock()
				if !isPong {
					sc.ChangeStream <- change

					pingPongSyncMapMutex.Lock()
					delete(pingPongSyncMap, change.AbsPath)
					pingPongSyncMapMutex.Unlock()
				}
      }
      time.Sleep(sc.fileWatcherSyncFreq)
    }
  }()
	return
}


func (sc *syncClient)AddDir(dir string, syncGroup int, errorStream chan error) {
	clientOrigin := rand.Intn(1000)

	// all changes from the server instance are passed to the change listener to prohibit the change listener from detecting it as "user side" change
	pingPongSync := make(chan util.FileChange)

	sc.ChangeListener(dir, pingPongSync, syncGroup, clientOrigin, errorStream)
	sc.SyncFromRemote(dir, pingPongSync, syncGroup, clientOrigin, errorStream)
}
