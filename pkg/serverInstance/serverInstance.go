package serverInstance

import (
	"goMinSync/pkg/remoteCacheToGo/cmd/cacheDb"
)

var dbPort = 8000

type serverInstance struct {
	cache cacheDb.CacheDb
}

func New() serverInstance {
	return serverInstance { cacheDb.New() }
}

func (instance serverInstance) StartInstance() {
	instance.cache.NewCache("changeLog")

	instance.cache.Db["changeLog"].RemoteConnHandler(8000)
}
