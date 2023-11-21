package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func getEntityAddrs() []string {
	hostname, err := os.Hostname()
	if err != nil {
		println("Error getting hostname: ", err.Error())
		os.Exit(0)
	}

	ips, _ := net.LookupIP(hostname)

	ipStrings := make([]string, len(ips))
	for i, ip := range ips {
		ipStrings[i] = ip.String() + ":5000"
	}

	return ipStrings
}

func prepSendAddr(addrs []string) []*net.UDPAddr {
	sendAddrs := make([]*net.UDPAddr, 0)
	for _, addr := range addrs {
		sendAddr, err := applyMask(addr)
		if err != nil {
			println("Error masking and resolving address: ", err.Error())
			os.Exit(0)
		}
		sendAddrs = append(sendAddrs, sendAddr)
	}
	return sendAddrs
}

func createSendingList(exception *net.UDPAddr, list []*net.UDPAddr) []*net.UDPAddr {
	sendList := make([]*net.UDPAddr, 0)
	exceptionMasked, err := applyMask(exception.String())
	if err != nil {
		println("Error masking exception address: ", err.Error())
		os.Exit(0)
	}
	for _, addr := range list {
		if addr.String() != exceptionMasked.String() {
			sendList = append(sendList, addr)
		}
	}
	return sendList
}

func applyMask(addr string) (*net.UDPAddr, error) {
	parts := strings.Split(addr, ".")
	parts[2] = "255"
	parts[3] = "255:5000"
	return net.ResolveUDPAddr("udp", strings.Join(parts, "."))
}

func decode(buffer []byte) (string, int, string) {
	var source, dest string

	for i := 0; i < 4; i++ {
		if i >= 3 {
			source += fmt.Sprintf("%02X", buffer[i])
		} else {
			source += fmt.Sprintf("%02X", buffer[i]) + ":"
		}
	}

	for i := 5; i < 9; i++ {
		if i >= 8 {
			dest += fmt.Sprintf("%02X", buffer[i])
		} else {
			dest += fmt.Sprintf("%02X", buffer[i]) + ":"
		}
	}
	return source, int(buffer[4]), dest
}
