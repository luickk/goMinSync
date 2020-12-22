package main

import (
	"fmt"
	"time"

  "remoteCacheToGo/cmd/cacheClient"
)


func main() {
	// unencryptd (for testing purposes)
	client, err := cacheClient.New("127.0.0.1", 8000, false, "", "")
  if err != nil {
    fmt.Println(err)
    return
  }
  var latestChange string
  for {
    latestChange = string(client.GetValByIndex(0))
    if latestChange != "" {
      fmt.Println(latestChange)
    }
    time.Sleep(10 * time.Millisecond)
  }
}
