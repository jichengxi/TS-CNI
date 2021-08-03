package main

import (
	"fmt"
	"log"
	"ts-cni/cni/utils"
)

func main() {
	K8sClient := utils.NewK8s()
	fmt.Println(K8sClient)
	netArr := K8sClient.GetPodNet("default", "nginx-test-5f6cc55c7f-jxrsq")
	log.Println("CNI NetArr的值=", netArr)
}
