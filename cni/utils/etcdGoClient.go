package utils

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"log"
	"strings"
	"time"
)

type Client struct {
	cli clientv3.Client
}

type EtcdGetValue struct {
	K string
	V string
}

// EtcdConnect etcd建立连接
func (c *Client) EtcdConnect() {
	config := clientv3.Config{
		//Endpoints:   []string{"192.168.1.25:2379"},
		Endpoints:   []string{"172.17.47.201:2379"},
		DialTimeout: 5 * time.Second,
	}
	cli, err := clientv3.New(config)
	if err != nil {
		fmt.Println("connect to etcd failed,err:", err)
		return
	}
	c.cli = *cli
}

// EtcdPut 创建和更新键值
func (c *Client) EtcdPut(k string, v string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	putRes, err := c.cli.Put(ctx, k, v, clientv3.WithPrevKV())
	cancel()
	if err != nil {
		log.Println("EtcdPut failed err:", err, "shit!")
		return
	}
	log.Println("put的上一次值", putRes.PrevKv)
}

// EtcdGet 通过isDir来控制get目录还是get具体的值
func (c *Client) EtcdGet(k string, isDir bool) interface{} {
	if isDir {
		// 这个条件说明需要查询目录
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		getRes, err := c.cli.Get(ctx, k, clientv3.WithKeysOnly(), clientv3.WithPrefix())
		cancel()
		if err != nil {
			log.Println("EtcdGet failed,err:", err, "shit!")
			return nil
		}
		var dirSlice []string
		for _, ev := range getRes.Kvs {
			tmpStr := strings.Split(strings.TrimSpace(string(ev.Key)), k)
			dirSlice = append(dirSlice, tmpStr[len(tmpStr) -1])
		}
		return dirSlice
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		getRes, err := c.cli.Get(ctx, k)
		cancel()
		if err != nil {
			log.Println("EtcdGet failed,err:", err, "shit!")
			return nil
		}
		var kvSlice []EtcdGetValue
		for _, ev := range getRes.Kvs {
			tmpGet := EtcdGetValue{K: string(ev.Key), V: string(ev.Value)}
			kvSlice = append(kvSlice, tmpGet)
		}
		return kvSlice
	}
}

// EtcdDel 删除一个键
func (c *Client) EtcdDel(k string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	delResp, err := c.cli.Delete(ctx, k)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("EtcdDel response：", delResp)
}

// EtcdDisconnect 关闭连接
func (c *Client)EtcdDisconnect() {
	c.cli.Close()
}
