package utils

import (
	"log"
	"strings"
	"ts-cni/cni/structs"
)

// MakeRange 创建一个连续数字的数组
func MakeRange(minNum int, mixNum int) []int {
	j := make([]int, mixNum-minNum+1)
	for i := range j {
		j[i] = minNum + i
	}
	return j
}

// Difference 求差集必须在其中一个数组是并集的基础上
func Difference(slice1 []string, slice2 []string) []string {
	//a是并集
	a := slice1
	b := slice2
	if len(a) > len(b) {
		a, b = b, a
	}
	m := make(map[string]int)
	n := make([]string, 0)
	for _, v := range a {
		m[v]++
	}
	for _, value := range b {
		times, _ := m[value]
		if times == 0 {
			n = append(n, value)
		}
	}
	return n
}

// EtcdAddIp 查询etcd现有ip使用列表,并新增IP
func EtcdAddIp(etcdClient *EtcdClient, netArr []string) structs.NetInfo {
	var resNetInfo structs.NetInfo
	// 取etcd中存储的所有的网段
	// 根据Annotations中的app_net
	// 第一步 查询app_net是否在etcd所有网段中存在
	// 第二步 如果存在判断app_net的整个网段的IP在etcd中是否超过了240个(11-250)
	// 第三步 如果超过了，就取app_net中下一个网段，如果没有下一个网段，报地址池IP不够
	// 第四步 如果没超过240个，就在etcd中取当前网段的vlanID，并且拼接到master上
	var etcdRootDir = "/ipam"
	netAllList := etcdClient.EtcdGet(etcdRootDir, true).([]string)
	for i, v := range netArr {
		if IsExistString(v, netAllList) {
		UsedIpList:
			usedIpList := etcdClient.EtcdGet(etcdRootDir+"/"+v, true).([]string)
			if len(usedIpList) < 240 {
				netVlanId := etcdClient.EtcdGet(etcdRootDir+"/"+v, false).([]EtcdGetValue)[0].V
				//n.Master = n.Master + "." + netVlanId
				//n.NetInfo.AppNet = v
				//n.NetInfo.UseIpList = usedIpList
				resNetInfo.AppNet = netVlanId
				resNetInfo.AppNet = v
				resNetInfo.UseIpList = usedIpList
				ResIp(etcdClient, &resNetInfo)
				if resNetInfo.IPAddress != "" && resNetInfo.GateWay != "" {
					return resNetInfo
				} else {
					goto UsedIpList
				}
			} else if i == len(usedIpList)-1 {
				log.Println("Annotations里的app_net地址池不够了!")
				return resNetInfo
			} else {
				log.Printf("%v 这个IP地址段中已经没有IP了! \n", v)
				continue
			}
		} else {
			log.Println("Annotations里的app_net写的有问题，找不到!")
			continue
		}
	}
	return resNetInfo
}

// EtcdDelIp 删除解锁并删除现有的ip的key
func EtcdDelIp(etcdClient *EtcdClient, podNs string, podName string) {
	k8sClient := NewK8s()
	podIp := k8sClient.GetPodIp(podNs, podName)
	// 获取ip前缀
	abcIpRange := podIp[:strings.LastIndex(podIp, ".")]
	var etcdRootDir = "/ipam"
	leaseId := etcdClient.EtcdGet(etcdRootDir+"/"+abcIpRange+"0/"+podIp, false).([]EtcdGetValue)[0].V
	err := etcdClient.UnLock(leaseId)
	if err != nil {
		log.Println("删除失败, err=", err)
	} else {
		isExist := etcdClient.EtcdGet(etcdRootDir+"/"+abcIpRange+"0/"+podIp, true).([]string)[0]
		if isExist == "" {
			log.Println(etcdRootDir+"/"+abcIpRange+"0/"+podIp, "真的删掉了!")
		}
	}
}

// EtcdCmdAdd 连接etcd进行操作
func EtcdCmdAdd(netArr []string) structs.NetInfo {
	// 建立ETCD连接
	etcdClient := EtcdClient{}
	etcdClient.EtcdConnect()
	defer etcdClient.EtcdDisconnect()
	ipNetInfo := EtcdAddIp(&etcdClient, netArr)
	return ipNetInfo
}

func EtcdCmdDel(podNs string, podName string) {
	// 建立ETCD连接
	etcdClient := EtcdClient{}
	etcdClient.EtcdConnect()
	defer etcdClient.EtcdDisconnect()
	EtcdDelIp(&etcdClient, podNs, podName)
}
