// +build linux

package node

import (
	"fmt"
	"net"

	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/config"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/util"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	kapi "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	utilnet "k8s.io/utils/net"
)

// getDefaultGatewayInterfaceDetails returns the interface name on
// which the default gateway (for route to 0.0.0.0) is configured.
// It also returns the default gateways themselves.
func getDefaultGatewayInterfaceDetails() (string, []net.IP, error) {
	var intfName string
	var gatewayIPs []net.IP

	if config.IPv4Mode {
		intfIPv4Name, gw, err := getDefaultGatewayInterfaceByFamily(netlink.FAMILY_V4)
		if err != nil {
			return "", gatewayIPs, err
		}
		intfName = intfIPv4Name
		gatewayIPs = append(gatewayIPs, gw)
	}

	if config.IPv6Mode {
		intfIPv6Name, gw, err := getDefaultGatewayInterfaceByFamily(netlink.FAMILY_V6)
		if err != nil {
			return "", gatewayIPs, err
		}
		// validate that both IP Families use the same interface for the gateway
		if intfName == "" {
			intfName = intfIPv6Name
		} else if intfName != intfIPv6Name {
			return "", nil, fmt.Errorf("multiple gateway interfaces detected: %s %s", intfName, intfIPv6Name)
		}
		gatewayIPs = append(gatewayIPs, gw)
	}

	if len(gatewayIPs) == 0 {
		return "", nil, fmt.Errorf("failed to get default gateway interface")
	}
	return intfName, gatewayIPs, nil
}

func getDefaultGatewayInterfaceByFamily(family int) (string, net.IP, error) {
	// filter the default route to obtain the gateway
	filter := &netlink.Route{Dst: nil}
	routes, err := netlink.RouteListFiltered(family, filter, netlink.RT_FILTER_DST)
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to get routing table in node")
	}
	// only support ECMP
	if len(routes) != 1 {
		return "", nil, fmt.Errorf("multiple default routes not allowed, use equal cost multipath instead")
	}
	r := routes[0]
	// no multipath
	if len(r.MultiPath) == 0 {
		intfLink, err := netlink.LinkByIndex(r.LinkIndex)
		if err != nil {
			return "", nil, errors.Wrapf(err, "failed to get interface link for route %v", r)
		}
		klog.Infof("Found default gateway %s %s", intfLink.Attrs().Name, r.Gw.String())
		return intfLink.Attrs().Name, r.Gw, nil
	}

	// multipath, use the first valid entry
	// TODO: revisit for full multipath support
	// xref: https://github.com/vishvananda/netlink/blob/6ffafa9fc19b848776f4fd608c4ad09509aaacb4/route.go#L137-L145
	for _, nh := range r.MultiPath {
		intfLink, err := netlink.LinkByIndex(nh.LinkIndex)
		if err != nil {
			klog.Warningf("Failed to get Link information: %v", err)
			continue
		}
		klog.Infof("Found default gateway %s %s", intfLink.Attrs().Name, nh.Gw.String())
		return intfLink.Attrs().Name, nh.Gw, nil
	}

	return "", net.IP{}, fmt.Errorf("failed to get default gateway interface")
}

func getDefaultIfAddr(defaultGatewayIntf string) (*net.IPNet, *net.IPNet, error) {
	var v4IfAddr, v6IfAddr *net.IPNet
	primaryLink, err := netlink.LinkByName(defaultGatewayIntf)
	if err != nil {
		return nil, nil, fmt.Errorf("error: unable to get link for default interface: %s, err: %v", defaultGatewayIntf, err)
	}
	addrs, err := netlink.AddrList(primaryLink, netlink.FAMILY_ALL)
	if err != nil {
		return nil, nil, fmt.Errorf("error: unable to list addresses for default interface, err: %v", err)
	}
	for _, addr := range addrs {
		if addr.IP.IsGlobalUnicast() {
			if utilnet.IsIPv6(addr.IP) {
				v6IfAddr = addr.IPNet
			} else {
				v4IfAddr = addr.IPNet
			}
		}
	}
	return v4IfAddr, v6IfAddr, nil
}

func getIntfName(gatewayIntf string) (string, error) {
	// The given (or autodetected) interface is an OVS bridge and this could be
	// created by us using util.NicToBridge() or it was pre-created by the user.

	// Is intfName a port of gatewayIntf?
	intfName, err := util.GetNicName(gatewayIntf)
	if err != nil {
		return "", err
	}
	_, stderr, err := util.RunOVSVsctl("get", "interface", intfName, "ofport")
	if err != nil {
		return "", fmt.Errorf("failed to get ofport of %s, stderr: %q, error: %v",
			intfName, stderr, err)
	}
	return intfName, nil
}

func deleteConntrack(ip string, port int32, protocol kapi.Protocol) error {
	return util.DeleteConntrack(ip, port, protocol)
}
