// Copyright 2015 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test1

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/containernetworking/plugins/plugins/ipam/host-local/backend/allocator"
	"github.com/containernetworking/plugins/plugins/ipam/host-local/backend/disk"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	//current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
)

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("host-local"))
}

func parseResolvConf(filename string) (*types.DNS, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	dns := types.DNS{}
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip comments, empty lines
		if len(line) == 0 || line[0] == '#' || line[0] == ';' {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "nameserver":
			dns.Nameservers = append(dns.Nameservers, fields[1])
		case "domain":
			dns.Domain = fields[1]
		case "search":
			dns.Search = append(dns.Search, fields[1:]...)
		case "options":
			dns.Options = append(dns.Options, fields[1:]...)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &dns, nil
}

func cmdCheck(args *skel.CmdArgs) error {

	ipamConf, _, err := allocator.LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	// Look to see if there is at least one IP address allocated to the container
	// in the data dir, irrespective of what that address actually is
	store, err := disk.New(ipamConf.Name, ipamConf.DataDir)
	if err != nil {
		return err
	}
	defer store.Close()

	containerIpFound := store.FindByID(args.ContainerID, args.IfName)
	if containerIpFound == false {
		return fmt.Errorf("host-local: Failed to find address added by container %v", args.ContainerID)
	}

	return nil
}

func cmdAdd(args *skel.CmdArgs) error {
	ipamConf, confVersion, err := allocator.LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}
	log.Println("ipamConf的值=", ipamConf)
	log.Println("confVersion的值=", confVersion)

	result := &current.Result{CNIVersion: current.ImplementedSpecVersion}

	if ipamConf.ResolvConf != "" {
		dns, err := parseResolvConf(ipamConf.ResolvConf)
		if err != nil {
			return err
		}
		result.DNS = *dns
	}

	store, err := disk.New(ipamConf.Name, ipamConf.DataDir)
	if err != nil {
		return err
	}
	defer store.Close()

	// Keep the allocators we used, so we can release all IPs if an error
	// occurs after we start allocating
	allocs := []*allocator.IPAllocator{}

	// Store all requested IPs in a map, so we can easily remove ones we use
	// and error if some remain
	requestedIPs := map[string]net.IP{} //net.IP cannot be a key

	for _, ip := range ipamConf.IPArgs {
		requestedIPs[ip.String()] = ip
	}

	for idx, rangeset := range ipamConf.Ranges {
		allocator := allocator.NewIPAllocator(&rangeset, store, idx)

		// Check to see if there are any custom IPs requested in this range.
		var requestedIP net.IP
		for k, ip := range requestedIPs {
			if rangeset.Contains(ip) {
				requestedIP = ip
				delete(requestedIPs, k)
				break
			}
		}

		ipConf, err := allocator.Get(args.ContainerID, args.IfName, requestedIP)
		if err != nil {
			// Deallocate all already allocated IPs
			for _, alloc := range allocs {
				_ = alloc.Release(args.ContainerID, args.IfName)
			}
			return fmt.Errorf("failed to allocate for range %d: %v", idx, err)
		}

		allocs = append(allocs, allocator)

		result.IPs = append(result.IPs, ipConf)
	}

	// If an IP was requested that wasn't fulfilled, fail
	if len(requestedIPs) != 0 {
		for _, alloc := range allocs {
			_ = alloc.Release(args.ContainerID, args.IfName)
		}
		errstr := "failed to allocate all requested IPs:"
		for _, ip := range requestedIPs {
			errstr = errstr + " " + ip.String()
		}
		return fmt.Errorf(errstr)
	}

	result.Routes = ipamConf.Routes
	log.Println("result的值=", *result)
	// result的值=
	//{1.0.0 []
	//[ {Interface:<nil> Address:{IP:192.168.165.22 Mask:ffffff00} Gateway:192.168.165.2}] [{Dst:{IP:0.0.0.0 Mask:00000000} GW:<nil>}] {[]  [] []}}
	return types.PrintResult(result, confVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	ipamConf, _, err := allocator.LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	store, err := disk.New(ipamConf.Name, ipamConf.DataDir)
	if err != nil {
		return err
	}
	defer store.Close()

	// Loop through all ranges, releasing all IPs, even if an error occurs
	var errors []string
	for idx, rangeset := range ipamConf.Ranges {
		ipAllocator := allocator.NewIPAllocator(&rangeset, store, idx)

		err := ipAllocator.Release(args.ContainerID, args.IfName)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}

	if errors != nil {
		return fmt.Errorf(strings.Join(errors, ";"))
	}
	return nil
}
