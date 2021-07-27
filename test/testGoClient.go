package main

import (
	"fmt"
	"ts-cni/cni2/utils"
)

func main() {
	a := utils.NewK8s()
	//fmt.Println(*a.Client)
	b := a.GetPodNet("default", "nginx-test-847b659596-cfzcl")
	fmt.Println(b)
}
