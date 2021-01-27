package main

import (
	"fmt"
	"goMinSync/serverInstance"
)


func main() {
	// params: serverRoot, maxUploadSize, token, address, syncPort, fileServerPort, tlsCert, tlsKey
	instance := serverInstance.New("/Users/luickklippel/Documents/projekte/go/src/goMinSync/test/temp", 10000, "test", "127.0.0.1", 8000, 8001, "", "")

	if err := instance.StartInstance(); err != nil {
		fmt.Println(err)
	}
}
