package main

import (
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
	"log"
	"net"
)

type Range struct {
	RangeStart net.IP      `json:"rangeStart,omitempty"` // The first ip, inclusive
	RangeEnd   net.IP      `json:"rangeEnd,omitempty"`   // The last ip, inclusive
	Subnet     types.IPNet `json:"subnet"`
	Gateway    net.IP      `json:"gateway,omitempty"`
}

type RangeSet []Range

type IPAMConfig struct {
	*Range
	Name       string
	Type       string         `json:"type"`
	Routes     []*types.Route `json:"routes"`
	DataDir    string         `json:"dataDir"`
	ResolvConf string         `json:"resolvConf"`
	Ranges     []RangeSet     `json:"ranges"`
	IPArgs     []net.IP       `json:"-"` // Requested IPs from CNI_ARGS and args
}

type IPAMArgs struct {
	IPs []net.IP `json:"ips"`
}

type Net struct {
	Name          string      `json:"name"`
	CNIVersion    string      `json:"cniVersion"`
	IPAM          *IPAMConfig `json:"ipam"`
	RuntimeConfig struct {    // The capability arg
		IPRanges []RangeSet `json:"ipRanges,omitempty"`
	} `json:"runtimeConfig,omitempty"`
	Args *struct {
		A *IPAMArgs `json:"cni"`
	} `json:"args"`
}

func loadIpamConf(bytes []byte, envArgs string) (*IPAMConfig, string, error) {
	// byte
	// "cniVersion":"0.3.1",
	//"ipMasq":false,
	//"ipam":{
	//       "gateway":"192.168.165.2",
	//       "rangeEnd":"192.168.165.29",
	//       "rangeStart":"192.168.165.21",
	//       "routes":[{"dst":"0.0.0.0/0"}],
	//       "subnet":"192.168.16 5.0/24",
	//       "type":"my-host-local"},
	//"isGateway":true,
	//"master":"enp0s3",
	//"mode":"bridge",
	//"name":"macvlannet","
	//type":"mymacvlan-2"}
	n := Net{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, "", fmt.Errorf("failed to load netconf: %v", err)
	}

	return nil, "", nil
}

func cmdAdd(args *skel.CmdArgs) error {
	log.Println("args的值=", *args)
	//	ContainerID e2e296b3ef03ce531592cafcdc2bfb4e017622419b3dfc60ab64c7b27b9fffb5
	//	Netns       /proc/10513/ns/net
	//	IfName      eth0
	//	Args        IgnoreUnknown=1;K8S_POD_NAMESPACE=default;K8S_POD_NAME=nginx-test-7ff7b6476d-4stpg;K8S_POD_INFRA_CONTAINER_ID=e2e296b3ef03ce531592cafcdc2bfb4e017622419b3dfc60ab64c7b27b9fffb5
	//  Path        /opt/cni/bin
	//  StdinData   {"cniVersion":"0.3.1","ipMasq":false,"ipam":{"gateway":"192.168.165.2","rangeEnd":"192.168.165.29","rangeStart":"192.168.165.21","routes":[{"dst":"0.0.0.0/0"}],"subnet":"192.168.165.0/24","type":"my-host-local"},"isGateway":true,"master":"enp0s3","mode":"bridge","name":"macvlannet","type":"mymacvlan-2"}
	ipamConf, confVersion, err := loadIpamConf(args.StdinData, args.Args)
	if err != nil {
		return fmt.Errorf("loadIpamConf函数加载配置出现问题: %v", err)
	}

	result := &current.Result{}

	// { [] [{Version:4 Interface:<nil> Address:{IP:192.168.165.29 Mask:ffffff00} Gateway:192.168.165.2}] [{Dst:{IP:0.0.0.0 Mask:00000000} GW:<nil>}] {[]  [] []}}
	/*
		CNIVersion ""
		Interfaces []
		IPs [{Version:4 Interface:<nil> Address:{IP:192.168.165.29 Mask:ffffff00} Gateway:192.168.165.2}]
		Routes [{Dst:{IP:0.0.0.0 Mask:00000000} GW:<nil>}]
		DNS {[]  [] []}}
	*/
	return types.PrintResult(result, confVersion)
}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}

func cmdDel(args *skel.CmdArgs) error {
	return nil
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("tc-ipam"))
}
