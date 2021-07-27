package main

import (
	"fmt"
	"ts-cni/cni2/utils"
)

func main() {
	c := "172.17.11.0"
	//var b []string
	//b = strings.Split(strings.TrimSpace(a),"/ipam/")
	//fmt.Println(b[len(b)-1])
	a := utils.Client{}
	a.EtcdConnect()
	//b := a.EtcdGet("/ipam/", true)
	str := a.EtcdGet("/ipam/"+c, false)
	a.EtcdDisconnect()
	//if utils.IsExistString(c, b.([]string)) {
	//	str := a.EtcdGet("/ipam/"+c, false)
	//	fmt.Println(str)
	//}
	fmt.Println(str)
	//fmt.Println("/ipam/"+c)
}
