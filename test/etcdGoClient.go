package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	//"log"
	"time"
)

func main() {
	// 新建etcd/v3的client，连接本地的etcd-server
	cli, err := clientv3.New(clientv3.Config{
		// 节点信息
		Endpoints: []string{"192.168.1.25:2379"},
		// 设置超时时间为5秒
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("connect to etcd failed,err:", err)
		return
	}
	defer cli.Close()

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//putResp, err := cli.Put(ctx, "test", "google", clientv3.WithPrevKV())
	//fmt.Println(putResp.PrevKv)
	//cancel()
	//if err != nil {
	//	fmt.Println("put failed err:", err)
	//	return
	//}
	//fmt.Println("PUT aaa=123456789,response：", putResp)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, "/app01/192.168.15", clientv3.WithKeysOnly(), clientv3.WithPrefix())
	cancel()
	if err != nil {
		fmt.Println("get failed,err:", err)
		return
	}
	fmt.Println(resp.Kvs)
	//for _, ev := range resp.Kvs {
	//	fmt.Printf("key=%s val=%s", ev.Key, ev.Value)
	//}

	//// del操作
	//ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	//delResp, err := cli.Delete(ctx, "test")
	//cancel()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println("Delete aaa,response：", delResp)

}
