// The systemd-networkd-dns command obtains the DHCP DNS server via DBus.
//
// For this to work, you must be running both systemd-networkd and
// systemd-resolved and provide the network interface name for your NIC
// that's connected to the DHCP network.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/godbus/dbus"
)

func main() {
	args := os.Args
	flag.Usage = func() {
		fmt.Printf(`usage: %s <interface>

The systemd-networkd-dns command obtains the DHCP DNS server via DBus.

For this to work, you must be running both systemd-networkd and
systemd-resolved and provide the network interface name for your NIC
that's connected to the DHCP network.
`, args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(args) < 2 {
		log.Println("error: must provide interface name for DHCP-enabled NIC")
		flag.Usage()
		os.Exit(2)
	}

	dns, err := getDHCPDNSForInterfaceFromDBus(args[1])
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(dns)
}

func getDHCPDNSForInterfaceFromDBus(iface string) (string, error) {

	// First, we need to determine the index of the interface
	// for this given interface name.  Indicies start at 1.
	i, err := net.InterfaceByName(iface)
	if err != nil {
		return "", fmt.Errorf("net.InterfaceByName() failed: %v", err)
	}

	// Connect to the system DBus
	conn, err := dbus.SystemBus()
	if err != nil {
		return "", fmt.Errorf("failed to connect to system DBus: %v", err)
	}

	var linkPath dbus.ObjectPath
	var callFlags dbus.Flags

	netO := conn.Object("org.freedesktop.resolve1", "/org/freedesktop/resolve1")
	netO.Call("org.freedesktop.resolve1.Manager.GetLink", callFlags, i.Index).Store(&linkPath)

	linkO := conn.Object("org.freedesktop.resolve1", linkPath)
	variant, err := linkO.GetProperty("org.freedesktop.resolve1.Link.DNS")
	if err != nil {
		return "", fmt.Errorf("error fetching DNS property from DBus: %v", err)
	}

	var variantVal [][]interface{}
	variantVal = variant.Value().([][]interface{})

	// Check the IP version of the nameserver address that was returned
	//  2 == AF_INET,  26 == AF_INET6
	if variantVal[0][0].(int32) != 2 {
		return "", fmt.Errorf("IPv6 nameserver addresses are not currently supported")

	}

	ipVariantBytes := variantVal[0][1].([]uint8)

	s := make([]string, len(ipVariantBytes))

	for v := range ipVariantBytes {
		s[v] = strconv.Itoa(int(ipVariantBytes[v]))
	}
	return strings.Join(s, "."), nil
}
