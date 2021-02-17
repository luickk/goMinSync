package serverInstance

import (
	"fmt"
	"crypto/tls"
	"net/http"
	"strconv"

	"goMinSync/pkg/remoteCacheToGo/cache"
	"goMinSync/pkg/fileUploadServer"
)

type serverInstance struct {
	cache cache.Cache
	fileServer fileUploadServer.Server
	serverRoot string
	maxUploadSize int64
	token string
	address string
	syncPort int
	fileServerPort int
	tlsCert string
	tlsKey string
}

func New(serverRoot string, maxUploadSize int64, token string, address string, syncPort int, fileServerPort int, tlsCert string, tlsKey string, errorStream chan error) serverInstance {
	var tokenEnabled bool
	if tlsCert == "" && tlsKey == "" {
		tokenEnabled = false
	} else {
		tokenEnabled = true
	}
	return serverInstance {
		cache.New(errorStream),
		fileUploadServer.NewServer(serverRoot, maxUploadSize, tokenEnabled, token, false),
		serverRoot,
		maxUploadSize,
		token,
		address,
		syncPort,
		fileServerPort,
		tlsCert,
		tlsKey,
	}
}

func (instance serverInstance) StartInstance(errorStream chan error) {
	http.Handle("/upload", instance.fileServer)
	http.Handle("/files/", instance.fileServer)

	if instance.tlsCert != "" && instance.tlsCert != "" {
		// params: port int, bindAddress string, pwHash string, dosProtection bool, serverCert string, serverKey string
		go instance.cache.RemoteTlsConnHandler(instance.syncPort, instance.address, instance.token, false, instance.tlsCert, instance.tlsKey, errorStream)

		go func() {
			cert, err := tls.X509KeyPair([]byte(instance.tlsCert), []byte(instance.tlsKey))
			if err != nil {
				fmt.Println("err")
				errorStream <- err
				return
			}

			tlsConnServer := http.Server {
				Addr:      instance.address+":"+strconv.Itoa(instance.fileServerPort),
				TLSConfig: &tls.Config{
					Certificates: []tls.Certificate{cert},
				},
			}
			if err := tlsConnServer.ListenAndServeTLS("", ""); err != nil {
				fmt.Println("err")
				errorStream <- err
				return
			}
		}()
	} else {
		go instance.cache.RemoteConnHandler(instance.address, instance.syncPort, errorStream)
		go func() {
			if err := http.ListenAndServe(instance.address+":"+strconv.Itoa(instance.fileServerPort), nil); err != nil {
				fmt.Println("err")
				errorStream <- err
				return
			}
		}()
	}
	return
}
