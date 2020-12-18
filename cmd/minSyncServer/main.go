package main

import (
	"goMinSync/pkg/serverInstance"
)


func main() {
	instance := serverInstance.New()
	instance.StartInstance()
}
