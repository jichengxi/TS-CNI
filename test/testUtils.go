package main

import (
	"fmt"
	"ts-cni/cni2/utils"
)

func main() {
	a := utils.Client{}
	a.EtcdConnect()
	t1 := a.EtcdGet("/ipam/", true)
	//t2 := a.EtcdGet("/test01/test01", false)
	a.EtcdDisconnect()
	fmt.Println(t1)
	//fmt.Println(t2)
}
