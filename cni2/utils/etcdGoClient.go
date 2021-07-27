package utils

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"log"
	"time"
)

type Client struct {
	cli clientv3.Client
}

type EtcdGetValue struct {
	K string
	V string
}

func (c *Client) EtcdConnect() {
	config := clientv3.Config{
		Endpoints:   []string{"192.168.1.25:2379"},
		DialTimeout: 5 * time.Second,
	}
	cli, err := clientv3.New(config)
	if err != nil {
		fmt.Println("connect to etcd failed,err:", err)
		return
	}
	c.cli = *cli
}

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
			dirSlice = append(dirSlice, string(ev.Key))
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

func (c *Client) EtcdDel(k string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	delResp, err := c.cli.Delete(ctx, k)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("EtcdDel response：", delResp)
}

func (c *Client)EtcdDisconnect() {
	c.cli.Close()
}
