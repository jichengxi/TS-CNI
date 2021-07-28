package main

import (
	"ts-cni/cni/utils"
)

func main() {
	a := utils.Client{}
	a.EtcdConnect()
	//lock, err := a.Lock("/172.11.11.11")
	//if err != nil {
	//	fmt.Println("groutine1抢锁失败")
	//	fmt.Println(err)
	//	return
	//}
	//lock := clientv3.LeaseID(int64(112443675516528807))
	//a.UnLock(lock)

	a.EtcdDisconnect()
}
