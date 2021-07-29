package utils

import (
	"log"
	"strings"
	"ts-cni/cni/structs"
)

var dIpRange = MakeRange(11, 30)

// ResIp 分配IP，返回IP
func ResIp(etcdClient *EtcdClient, NetInfo *structs.NetInfo) {
	// 获取ip前缀
	abcIpRange := NetInfo.AppNet[:strings.LastIndex(NetInfo.AppNet, ".")]
	var ipAllList []string
	for i := 0; i < len(dIpRange)-1; i++ {
		// 该网段IP总列表
		ipAllList = append(ipAllList, abcIpRange+"."+string(rune(dIpRange[i])))
	}
	// 根据已使用的和总列表取出差集
	diffIpList := Difference(NetInfo.UseIpList, ipAllList)
	resIp := diffIpList[0]
	resGw := abcIpRange + ".254"
	err := etcdClient.Lock("/ipam/" + NetInfo.AppNet + "/" + resIp)
	if err != nil {
		log.Println("加锁失败, err=", err)
		NetInfo.IPAddress = ""
		NetInfo.GateWay = ""
	} else {
		NetInfo.IPAddress = resIp
		NetInfo.GateWay = resGw
	}
}
