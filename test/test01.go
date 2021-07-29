package main

import (
	"fmt"
	"ts-cni/cni/utils"
)

func main() {
	a := utils.EtcdClient{}
	a.EtcdConnect()
	b := a.EtcdGet("/i", false).(string)
	a.EtcdDisconnect()
	fmt.Println(b)
}
