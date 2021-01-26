package main

import (
  "fmt"

	"goMinSync/syncClient"
)


func main() {
	sC := syncClient.New()

  err := sC.ConnectToRemoteInstance("127.0.0.1", 8000)
  if err != nil {
    fmt.Println(err)
    return
  }

  // should be located beneath err return of syncClient connect
  sC.AddDir("/Users/luickklippel/Documents/Temp")

  sC.StartSyncToRemoteInstance()
  for {}
}
