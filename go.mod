module ts-cni

go 1.16

require (
	github.com/containernetworking/cni v0.8.1
	github.com/containernetworking/plugins v0.9.1
	github.com/j-keck/arping v0.0.0-20160618110441-2cf9dc699c56
	github.com/vishvananda/netlink v1.1.1-0.20201029203352-d40f9887b852
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489
	k8s.io/client-go v0.21.3
)

replace (
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)
