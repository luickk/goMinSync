package serverInstance

import (
	"crypto/tls"
	"net/http"
	"fmt"

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

func New(serverRoot string, maxUploadSize int64, token string, address string, syncPort int, fileServerPort int, tlsCert string, tlsKey string) serverInstance {
	return serverInstance { cache.New(), fileUploadServer.NewServer(serverRoot, maxUploadSize, token, false), serverRoot, maxUploadSize, token, address, syncPort, fileServerPort, tlsCert, tlsKey }
}

func (instance serverInstance) StartInstance() error {
	http.Handle("/upload", instance.fileServer)
	http.Handle("/files/", instance.fileServer)

	if instance.tlsCert != "" && instance.tlsCert != ""{
		go instance.cache.RemoteTlsConnHandler(instance.syncPort, instance.token, false, instance.tlsCert, instance.tlsKey)

		cert, err := tls.X509KeyPair([]byte(instance.tlsCert), []byte(instance.tlsKey))
		if err != nil {
			fmt.Println(err)
			return err
		}
		tlsConnServer := http.Server{
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		}
		if err := tlsConnServer.ListenAndServeTLS(fmt.Sprintf("%s:%d", instance.address, instance.fileServerPort), ""); err != nil {
			return err
		}
	} else {
		go instance.cache.RemoteConnHandler(instance.syncPort)
		if err := http.ListenAndServe(fmt.Sprintf("%s:%d", instance.address, instance.fileServerPort), nil); err != nil {
			return err
		}
	}
	return nil
}
