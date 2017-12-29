package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/godbus/dbus"
)

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

	var ipBytes []byte

	ipVariantBytes := variantVal[0][1].([]uint8)
	for _, v := range ipVariantBytes {
		ipBytes = append(ipBytes, byte(v))
	}
	return convertIPBytesToIPAddress(ipBytes), nil

}

func convertIPBytesToIPAddress(b []byte) string {
	s := make([]string, len(b))
	for i := range b {
		s[i] = strconv.Itoa(int(b[i]))
	}
	return strings.Join(s, ".")
}
