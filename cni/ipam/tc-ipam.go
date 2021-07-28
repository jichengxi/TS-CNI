package ipam

import (
	"strings"
	"ts-cni/cni/structs"
	"ts-cni/cni/utils"
)

var dIpRange = utils.MakeRange(11, 30)

func ResIp(NetInfo structs.NetInfo) (string, string) {
	// ip前缀
	abcIpRange := NetInfo.AppNet[:strings.LastIndex(NetInfo.AppNet, ".")]
	var ipAllList []string
	for i := 0; i < len(dIpRange)-1; i++ {
		// 该网段IP总列表
		ipAllList = append(ipAllList, abcIpRange+"."+string(rune(dIpRange[i])))
	}
	// 根据已使用的和总列表取出差集
	diffIpList := utils.Difference(NetInfo.UseIpList, ipAllList)
	resIp := diffIpList[0]
	resGw := abcIpRange + ".254"
	return resIp, resGw
}
