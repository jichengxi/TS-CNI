package main

import (
	"fmt"
	"time"
	"ts-cni/cni/utils"
)

func main() {
	a := utils.Client{}
	a.EtcdConnect()
	go func() {
		lock, err := a.Lock("/172.11.11.11")
		if err != nil {
			fmt.Println("groutine1抢锁失败")
			fmt.Println(err)
			return
		}
		fmt.Println("groutine1抢锁成功")
		fmt.Println("lease id", lock)
		time.Sleep(10 * time.Second)
	}()
	a.EtcdDisconnect()
	//a.UnLock(lock)

}
