package main

import (
	"log"
	"ts-cni/cni/utils"
)

type N struct {
	Master string
}

func main() {
	K8sClient := utils.NewK8s()
	netArr := K8sClient.GetPodNet("default", "nginx-test-847b659596-cfzcl")
	etcdClient := utils.Client{}
	etcdClient.EtcdConnect()
	// 取etcd中存储的所有的网段
	n := N{Master: "eth0"}
	var etcdRootDir = "/ipam"
	netList := etcdClient.EtcdGet(etcdRootDir, true).([]string)
	for _, v := range netArr {
		if utils.IsExistString(v, netList) {
			usedIpList := etcdClient.EtcdGet(etcdRootDir+"/"+v, true).([]string)
			if len(usedIpList) < 2 {
				netVlanId := etcdClient.EtcdGet(etcdRootDir+"/"+v, false).([]utils.EtcdGetValue)[0].V
				n.Master = n.Master + "." + netVlanId
				break
			} else {
				log.Printf("%v 这个IP地址段中已经没有IP了! \n", v)
				continue
			}
		} else {
			log.Printf("%v 的Annotations里的app_net写的有问题，找不到! \n", "nginx-test-7ff7b6476d-jbfqc")
		}
	}
	etcdClient.EtcdDisconnect()
	log.Println(n.Master)

}
