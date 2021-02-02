package main

import (
  "fmt"

	"goMinSync/syncClient"
)


func main() {
	sC := syncClient.New()

  err := sC.ConnectToRemoteInstance("127.0.0.1", 8000, 8001, "", "")
  if err != nil {
    fmt.Println(err)
    return
  }

  // should be located beneath err return of syncClient connect
  if err := sC.AddDir("/Users/luickklippel/Documents/Temp Local", 1); err != nil {
    fmt.Println(err)
    return
  }

  if err := sC.AddDir("/Users/luickklippel/Documents/test", 1); err != nil {
    fmt.Println(err)
    return
  }
  
  for {}
}
