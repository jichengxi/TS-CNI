package utils

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"log"
	"strconv"
	"strings"
	"time"
)

type EtcdGetValue struct {
	K string
	V string
}

// EtcdClient 分布式锁(TXN事务)
type EtcdClient struct {
	// etcd客户端
	cli        clientv3.Client
	Kv         clientv3.KV
	Lease      clientv3.Lease
	CancelFunc context.CancelFunc // 用于终止自动续租
	LeaseId    clientv3.LeaseID   // 租约ID
	IsLocked   bool               // 是否上锁成功
	txn        clientv3.Txn
}

// EtcdConnect etcd建立连接
func (c *EtcdClient) EtcdConnect() {
	config := clientv3.Config{
		//Endpoints: []string{"192.168.1.25:2379"},
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
func (c *EtcdClient) EtcdPut(k string, v string) {
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
func (c *EtcdClient) EtcdGet(k string, isDir bool) interface{} {
	if isDir {
		// 这个条件说明需要查询目录
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		// 判断是不是/结尾的，如果不是，就给他加上
		if []byte(k)[len([]byte(k))-1] != 47 {
			k = k + "/"
		}
		getRes, err := c.cli.Get(ctx, k, clientv3.WithKeysOnly(), clientv3.WithPrefix())
		cancel()
		if err != nil {
			log.Println("EtcdGet failed,err:", err, "shit!")
			return nil
		}
		var dirSlice []string
		for _, ev := range getRes.Kvs {
			tmpStrSlice := strings.Split(strings.TrimSpace(string(ev.Key)), k)
			tmpStr := tmpStrSlice[len(tmpStrSlice)-1]
			tmpByte := []byte(tmpStr)
			if !IsExistByte(byte(47), tmpByte) {
				dirSlice = append(dirSlice, tmpStr)
			}
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
func (c *EtcdClient) EtcdDel(k string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	delResp, err := c.cli.Delete(ctx, k)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("EtcdDel response：", delResp)
}

// EtcdDisconnect 关闭连接
func (c *EtcdClient) EtcdDisconnect() {
	err := c.cli.Close()
	if err != nil {
		return
	}
}

// InitLock 初始化锁
func (c *EtcdClient) InitLock() error {
	var ctx context.Context
	c.Kv = clientv3.NewKV(&c.cli)
	c.Lease = clientv3.NewLease(&c.cli)
	leaseResp, err := c.Lease.Grant(context.TODO(), 10)
	if err != nil {
		return err
	}
	ctx, c.CancelFunc = context.WithCancel(context.TODO())
	c.LeaseId = leaseResp.ID
	log.Println("锁内存地址:", &c.LeaseId)
	KeepResChan, err := c.Lease.KeepAlive(ctx, c.LeaseId)
	go func() {
		var keepRes *clientv3.LeaseKeepAliveResponse
		for {
			select {
			case keepRes = <-KeepResChan:
				//如果续约失败
				if keepRes != nil {
					log.Println("续租成功,leaseID :", keepRes.ID)
					log.Println("锁内存地址:", &keepRes.ID)
				} else {
					log.Println("续租失败")
					c.CancelFunc()
					return
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()
	return err
}

// Lock 创建锁
func (c *EtcdClient) Lock(key string) error {
	// 初始化锁，一般用于创建前
	err := c.InitLock()
	if err != nil {
		return err
	}
	txn := c.Kv.Txn(context.TODO())
	// 判断锁是否存在
	txn.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, strconv.FormatInt(int64(c.LeaseId), 10), clientv3.WithLease(c.LeaseId))).
		Else()
	txnResp, err := txn.Commit()
	if err != nil {
		return err
	}
	// 判断锁是否存在
	if !txnResp.Succeeded { //判断txn.if条件是否成立
		return fmt.Errorf("抢锁失败")
	}
	return nil
}

// UnLock 通过锁的字符串删除锁
func (c *EtcdClient) UnLock(leaseIdStr string) error {
	leaseId := c.SearchLocks(leaseIdStr)
	revoke, err := c.cli.Revoke(context.TODO(), leaseId)
	if err != nil {
		log.Printf("获取%v锁出错, %v", leaseIdStr, err)
		return err
	}
	log.Println("清除锁成功，返回值=", revoke.Header)
	return nil
}

// SearchLocks 通过锁字符串查询出具体锁的id
func (c *EtcdClient) SearchLocks(leaseIdStr string) clientv3.LeaseID {
	leases, err := c.cli.Leases(context.TODO())
	errId := clientv3.LeaseID(-000000000000)
	if err != nil {
		log.Println("获取所有锁出错,", err)
		return errId
	}
	log.Println(leases.Leases)
	if leases != nil {
		for _, i := range leases.Leases {
			if leaseIdStr == strconv.FormatInt(int64(i.ID), 10) {
				log.Println("锁内存地址:", &i.ID)
				log.Println("i=", i.ID)
				return i.ID
			}
		}
	} else {
		log.Println("没有锁了")
		return errId
	}
	return errId
}
