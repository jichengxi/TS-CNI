package main

import (
	"fmt"
	"ts-cni/cni2/utils"
)

func main() {
	a := utils.Client{}
	a.EtcdConnect()
	//a.EtcdPut("/test01", "1234")
	//a.EtcdPut("/test01/test01", "test123")
	//a.EtcdPut("/test01/test02", "test123")
	t1 := a.EtcdGet("/test01/", true)
	t2 := a.EtcdGet("/test01/test01", false)
	fmt.Println(t1)
	fmt.Println(t2)
}
