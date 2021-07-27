package main

import (
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
	_ "github.com/containernetworking/plugins/pkg/utils/sysctl"
	"github.com/vishvananda/netlink"
	"log"
	"net"
	"runtime"
)

type EnvArgs struct {
	types.CommonArgs
	K8sPodNamespace        string `json:"K8S_POD_NAMESPACE"`
	K8sPodName             string `json:"K8S_POD_NAME"`
	K8sPodInfraContainerId string `json:"K8S_POD_INFRA_CONTAINER_ID"`
}

type NetConf struct {
	types.NetConf
	Master  string `json:"master"`
	Mode    string `json:"mode"`
	MTU     int    `json:"mtu"`
	Mac     string `json:"mac,omitempty"`
	EnvArgs EnvArgs
}

//const (
//	IPv4InterfaceArpProxySysctlTemplate = "net.ipv4.conf.%s.proxy_arp"
//)

func init() {
	log.SetPrefix("TS-CNI: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	runtime.LockOSThread()
}

func getDefaultRouteInterfaceName() (string, error) {
	routeToDstIP, err := netlink.RouteList(nil, netlink.FAMILY_ALL)
	if err != nil {
		return "", err
	}

	for _, v := range routeToDstIP {
		if v.Dst == nil {
			l, err := netlink.LinkByIndex(v.LinkIndex)
			if err != nil {
				return "", err
			}
			return l.Attrs().Name, nil
		}
	}

	return "", fmt.Errorf("no default route interface found")
}

func getMTUByName(ifName string) (int, error) {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return 0, err
	}
	return link.Attrs().MTU, nil
}

func modeFromString(s string) (netlink.MacvlanMode, error) {
	switch s {
	case "", "bridge":
		return netlink.MACVLAN_MODE_BRIDGE, nil
	case "private":
		return netlink.MACVLAN_MODE_PRIVATE, nil
	case "vepa":
		return netlink.MACVLAN_MODE_VEPA, nil
	case "passthru":
		return netlink.MACVLAN_MODE_PASSTHRU, nil
	default:
		return 0, fmt.Errorf("unknown macvlan mode: %q", s)
	}
}

func loadConf(bytes []byte, envArgs string) (*NetConf, string, error) {
	n := &NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, "", fmt.Errorf("failed to load netconf: %v", err)
	}
	log.Println("第一步的n:", *n)
	log.Println("第一步的envArgs:", envArgs)
	// 加载命令行传进来的参数
	// 命令行传进来的参数=
	// IgnoreUnknown=1;
	// K8S_POD_NAMESPACE=default;
	// K8S_POD_NAME=nginx-test;
	// K8S_POD_INFRA_CONTAINER_ID=c9955ddd4f37e4822f4ddb198e1c4069fa4598720897a07b68f4114267285c12
	if envArgs != "" {
		log.Println("envArgs转换前的值=", envArgs)
		m := &EnvArgs{}
		if err := json.Unmarshal([]byte(envArgs), m); err != nil {
			return nil, "", fmt.Errorf("failed to load envArgs: %v", err)
		}
		n.EnvArgs = *m
		log.Println("envArgs转换后的值=", *m)
		log.Println("转换后n的值=", *n)
	}

	// 没有设置网卡就使用默认网卡
	if n.Master == "" {
		defaultRouteInterface, err := getDefaultRouteInterfaceName()
		if err != nil {
			return nil, "", err
		}
		n.Master = defaultRouteInterface
	}

	// 配置MTU，在不设置的情况下就是0 没有什么卵用
	masterMTU, err := getMTUByName(n.Master)
	if err != nil {
		return nil, "", err
	}
	if n.MTU < 0 || n.MTU > masterMTU {
		return nil, "", fmt.Errorf("invalid MTU %d, must be [0, master MTU(%d)]", n.MTU, masterMTU)
	}

	return n, n.CNIVersion, nil
}

func createMacvlan(conf *NetConf, ifName string, netns ns.NetNS) (*current.Interface, error) {
	macvlan := &current.Interface{}
	// 转化配置文件中的mode
	mode, err := modeFromString(conf.Mode)
	if err != nil {
		return nil, err
	}
	log.Println("macvlan的mode", mode)

	m, err := netlink.LinkByName(conf.Master)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup master %q: %v", conf.Master, err)
	}
	log.Println("macvlan的m", m)

	// 子接口网卡名
	// due to kernel bug we have to create with tmpName or it might
	// collide with the name on the host and error out
	tmpName, err := ip.RandomVethName()
	if err != nil {
		return nil, err
	}

	linkAttrs := netlink.LinkAttrs{
		MTU:         conf.MTU,
		Name:        tmpName,
		ParentIndex: m.Attrs().Index,
		Namespace:   netlink.NsFd(int(netns.Fd())),
	}

	if conf.Mac != "" {
		addr, err := net.ParseMAC(conf.Mac)
		if err != nil {
			return nil, fmt.Errorf("invalid args %v for MAC addr: %v", conf.Mac, err)
		}
		linkAttrs.HardwareAddr = addr
	}
	log.Println("linkAttrs.HardwareAddr=", linkAttrs.HardwareAddr)

	// 整合macvlan所需要的参数
	mv := &netlink.Macvlan{
		LinkAttrs: linkAttrs,
		Mode:      mode,
	}

	if err := netlink.LinkAdd(mv); err != nil {
		return nil, fmt.Errorf("failed to create macvlan: %v", err)
	}

	err = netns.Do(func(_ ns.NetNS) error {
		// TODO: duplicate following lines for ipv6 support, when it will be added in other places
		//ipv4SysctlValueName := fmt.Sprintf(IPv4InterfaceArpProxySysctlTemplate, tmpName)
		//if _, err := sysctl.Sysctl(ipv4SysctlValueName, "1"); err != nil {
		//	// remove the newly added link and ignore errors, because we already are in a failed state
		//	_ = netlink.LinkDel(mv)
		//	return fmt.Errorf("failed to set proxy_arp on newly added interface %q: %v", tmpName, err)
		//}

		err := ip.RenameLink(tmpName, ifName)
		// 如果改名没改成功就需要把前面创建的网卡 --- 删除
		if err != nil {
			_ = netlink.LinkDel(mv)
			return fmt.Errorf("failed to rename macvlan to %q: %v", ifName, err)
		}
		macvlan.Name = ifName

		// Re-fetch macvlan to get all properties/attributes
		contMacvlan, err := netlink.LinkByName(ifName)
		log.Println("contMacvlan的值=", contMacvlan)
		if err != nil {
			return fmt.Errorf("failed to refetch macvlan %q: %v", ifName, err)
		}
		macvlan.Mac = contMacvlan.Attrs().HardwareAddr.String()
		macvlan.Sandbox = netns.Path()
		log.Println("macvlan的值=", *macvlan)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return macvlan, nil
}

func cmdAdd(args *skel.CmdArgs) error {
	log.Println("args的值:", *args)
	n, cniVersion, err := loadConf(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	// 网络命名空间
	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", netns, err)
	}
	defer netns.Close()

	// macvlanInterface参数： {eth0 82:e1:18:79:a4:5d /proc/10491/ns/net}
	macvlanInterface, err := createMacvlan(n, args.IfName, netns)
	if err != nil {
		return err
	}
	log.Println("macvlanInterface参数：", *macvlanInterface)

	// Delete link if err to avoid link leak in this ns
	defer func() {
		if err != nil {
			netns.Do(func(_ ns.NetNS) error {
				return ip.DelLinkByName(args.IfName)
			})
		}
	}()

	result := &current.Result{
		CNIVersion: cniVersion,
		Interfaces: []*current.Interface{macvlanInterface},
	}

	return types.PrintResult(result, cniVersion)
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("tc-cni"))
}
