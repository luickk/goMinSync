package syncClient

import (
	"fmt"

  "remoteCacheToGo/cmd/cacheClient"
)

type clientInstance struct {
	cacheClient cacheClient.RemoteCache
}

func New(address string, port int) (clientInstance, error) {
	var cI clientInstance
	cClient, err := connectToRemoteCache(address, port)
	if err != nil {
		fmt.Println(err)
		return cI, err
	}
	cI.cacheClient = cClient
	return cI, nil
}

func connectToRemoteCache(address string, port int) (cacheClient.RemoteCache, error) {
	client, err := cacheClient.New(address, port, false, "", "")
  if err != nil {
    fmt.Println(err)
    return client, err
  }
	return client, err
}
