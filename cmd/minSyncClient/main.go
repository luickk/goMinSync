package main

import (
  "fmt"

	"goMinSync/pkg/syncClient"
)


func main() {
	sc, err := syncClient.New("127.0.0.1", 8000)

  // should be located beneath err return of syncClient connect
  sc.AddDir("/Users/luickklippel/Documents/Temp")
  for {
    select {
    case chg := <-sc.ChangeStream:
      fmt.Println(chg.Ctype + ":" + chg.DirPath)
    }
  }
  if err != nil {
    fmt.Println(err)
    return
  }

}
