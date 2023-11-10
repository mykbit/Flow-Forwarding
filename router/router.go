package main

import (
	"net"
	"os"
	"strings"
	"sync"
)

var wg sync.WaitGroup

func main() {

	entityAddrList := getEntityAddrs()
	for _, addr := range entityAddrList {
		println("Entity address: ", addr)
	}
	sendAddrList := prepSendAddr(entityAddrList)

	listenAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:5000")
	if err != nil {
		println("Failed to resolve listening address: ", err.Error())
		os.Exit(0)
	}

	socket, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		println("Failed to establish socket: ", err.Error())
		os.Exit(0)
	}

	wg.Add(1)
	go receiveData(socket, sendAddrList, entityAddrList)
	wg.Wait()
}

func receiveData(socket *net.UDPConn, sendAddrList []*net.UDPAddr, entityAddrList []string) {
	defer wg.Done()

	for {
		buffer := make([]byte, 65000)

		n, senderAddr, err := socket.ReadFromUDP(buffer)
		if err != nil {
			println("Error reading from client: ", err.Error())
			os.Exit(0)
		}
		if senderAddrStr := senderAddr.String(); senderAddrStr != entityAddrList[0] && senderAddrStr != entityAddrList[1] {
			println("Received data from ", senderAddr.String())

			dataBuffer := make([]byte, n)
			copy(dataBuffer, buffer[:n])

			go forwardData(socket, dataBuffer, senderAddr, sendAddrList)
		}
	}
}

func forwardData(socket *net.UDPConn, data []byte, senderAddr *net.UDPAddr, sendAddrList []*net.UDPAddr) {
	sendList := createSendingList(senderAddr, sendAddrList)
	for _, addr := range sendList {
		_, err := socket.WriteToUDP(data, addr)
		if err != nil {
			println("Error sending data: ", err.Error())
			continue
		} else {
			println("Sent data to ", addr.String())
		}
	}
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
			println("Added ", addr.String(), " to sending list")
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
