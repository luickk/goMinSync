package main

import (
	"fmt"

	"goMinSync/serverInstance"
)


func main() {
	errorStream := make(chan error)
	// // params: serverRoot, maxUploadSize, token, address, syncPort, fileServerPort, tlsCert, tlsKey
	// instance := serverInstance.New("/Users/luickklippel/Documents/projekte/go/src/goMinSync/test/temp", "test", "127.0.0.1", 8000, 8001, "", "", errorStream)
	// instance.StartInstance(errorStream)

	instance := serverInstance.New("/Users/luickklippel/Documents/projekte/go/src/goMinSync/test/temp", "", "127.0.0.1", 8000, 8001, "", "", errorStream)
	instance.StartInstance(errorStream)

	for {
		if err :=<- errorStream; err != nil {
			fmt.Println(err.Error())
		}
	}
}
