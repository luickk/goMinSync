package main

import (
  "fmt"

	"goMinSync/syncClient"
)

func main() {
cert := `-----BEGIN CERTIFICATE-----
MIIFPzCCAyegAwIBAgIJAI8+kdLEcyiUMA0GCSqGSIb3DQEBBQUAMFIxCzAJBgNV
BAYTAkNOMQwwCgYDVQQIDANQRUsxETAPBgNVBAcMCEJlaSBKaW5nMQ8wDQYDVQQK
DAZWTXdhcmUxETAPBgNVBAMMCEhhcmJvckNBMB4XDTIxMDIxMzEwNDQxMloXDTIy
MDIxMzEwNDQxMlowVzELMAkGA1UEBhMCQ04xDDAKBgNVBAgMA1BFSzERMA8GA1UE
BwwIQmVpIEppbmcxDzANBgNVBAoMBlZNd2FyZTEWMBQGA1UEAwwNSGFyYm9yTWFu
YWdlcjCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAMDihRY9v+hBm+T5
VJl/9SIfocwgCDiM/oGgftrH/hzk8yDI2ztzFFLbsmpE75yPkntSy8nl8wulRUQQ
k6Ih+ZXEPySctyY0skdIw9UAfSaokFm7lfKDSifN0pnr/MCTWDyPrC/Srzt2xDlg
MaBAoh95Q5QahSYpP3k6s0uS0nC8HU7EmqY7xgxBLtlOO8uvewfC5xs0KTA69sxJ
JkBdkUCobPcDFE4r7d4mWiOQf50/+j0TU/fwcytQ7Gk+bhXKgKYeiaZQLTcmiBwe
7+Llvig8gPgj4ftEwWm3gMGYWurd+os2YoMUtFNp8Hib9bV35tU1qvwc8gfS+3iv
oeTlr6qrY3s9i9+qvwsZ/nQHpzrrqANXquT1u2BFO1qwTTF9hQyFm+mnjuhFBISx
dx/OWkWHEnuNCnzcdlBsE2sWlEMOprdbCJFxex1uBlyo9cxbxhrfZiZGCOremLwv
UHwl5C6dIZhNfuizpOwClHXxNLMgNPhRruX2aPctmM8zGwcQiliKvsr7YuIlnYXe
BW9yTbFpt5Nh1liXk+yQJR+wOp/ofajXnVgkvbzjy7oReAmVCXYmLEtUFpWhItFd
dwr50sH3VIO8s2JFuMbCv/4KzHukvgsN5h0dMRvXwNjaA5YkLtDwFkLhrWOEX3Hj
Gr+iK/1aZv4LPKAEKu0cwz4KrZnXAgMBAAGjEzARMA8GA1UdEQQIMAaHBH8AAAEw
DQYJKoZIhvcNAQEFBQADggIBACkgBWK8CclDfbhyFQFSe0rwh1fmmVljn/1dQaJf
OpGgbfrl15g/05VK9tIV30sUYcDoYS8/ZStGuHl4wX3Y+lyTzmZsjPSTu/0wc6T5
PNHug3z5xi+N5xievhIllKJPywDpWKM7X+t9fPSeVapKbpUNr+lCH4hdOYLaVXjO
EanU/3lSEaDflNtywLrZ5xg8KHJ2AWfOCPoJmuLdkbcXFLSv86empYIBY8lEYJCb
svk6AmeJSo42oHaJ7equFzmBYOsEtP3cm3mos/9ui06mU+KQW0xdrwGGEyTisq6e
TplqiJ2iGLkh0qupuvRFYkKu08RAfkCAZHKKWUJA2QgNms/XM+b8ttGt+S5ydrUh
MB0VW1SNirDfAvCYXzxgWLbRk6CInThxJ+q3QDFqk843iOpd/YNfGD6jW9hO6aJK
DaUQbwsWWg4MAwvd8bsfJ9WwB/zWiQfwLLrR8oxAyt3PU0qI+5xaq9lp2Pl4naA6
lGoag61ek8G6oKIAw6XXUy+UKJ81wIA7qajCpOrXFFW3QgWhFU3wwyixFNFWUj/H
tDmsJUs9P/27U7uS2cQy+u1VTvJFRKAfpLG38AZeAA+ttoT/N8IOTBE7NhKtTt8j
ZmMKcSu5NdUBt5Iz10kM7AM+eyF77obzdN3ixB7XrpzflnF4yGjp2LlUhkTCpBCk
dx2R
-----END CERTIFICATE-----`

	sC := syncClient.New()

  err := sC.ConnectToRemoteInstance("127.0.0.1", 8000, 8001, "test", cert)
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
