package main

import (
  "fmt"

	"goMinSync/syncClient"
)

func main() {
  errorStream := make(chan error)

	sC := syncClient.New()

  sC.ConnectToRemoteInstance("127.0.0.1", 8000, 8001, "", "", errorStream)

  // should be located beneath err return of syncClient connect
  go sC.AddDir("/Users/luickklippel/Documents/Temp Local", 1, errorStream)

  go sC.AddDir("/Users/luickklippel/Documents/test", 1, errorStream)

	for {
		if err :=<- errorStream; err != nil {
			fmt.Println(err.Error())
		}
	}
}
