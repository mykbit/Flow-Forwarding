package main

import (
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {

	roleStr := os.Getenv("ROLE")

	entityAddr := getEntityAddr()

	entityAddrMasked := applyMask(entityAddr)
	listenAddrRaw := "0.0.0.0" + ":" + os.Getenv("PORT")
	broadcastAddrRaw := entityAddrMasked + ":" + os.Getenv("PORT")

	listenAddr, err := net.ResolveUDPAddr("udp", listenAddrRaw)
	if err != nil {
		println("Failed to resolve broadcast address: ", err.Error())
		os.Exit(0)
	}

	broadcastAddr, err := net.ResolveUDPAddr("udp", broadcastAddrRaw)
	if err != nil {
		println("Failed to resolve broadcast address: ", err.Error())
		os.Exit(0)
	}

	socket, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		println("Failed to establish broadcast connection: ", err.Error())
		os.Exit(0)
	}
	defer socket.Close()

	println("Client broadcast connection established")

	dataPath := os.Getenv("DATA_PATH")
	data, err := os.ReadDir(dataPath)
	if err != nil {
		println("Error reading data directory: ", err.Error())
		os.Exit(0)
	}

	canStream, _ := strconv.Atoi(os.Getenv("CAN_STREAM"))

	switch roleStr {
	case "client":
		wg.Add(2)
		go streamData(socket, data, dataPath, broadcastAddr)
		go receiveDataClient(socket, entityAddr)
		wg.Wait()
	case "server":
		wg.Add(1)
		go receiveDataServer(socket, entityAddr, canStream)
		wg.Wait()
	default:
		println("Invalid role: ", roleStr)
	}
}

func streamData(socket *net.UDPConn, data []os.DirEntry, dataPath string, addr *net.UDPAddr) {
	defer wg.Done()
	for i := 1; i <= len(data); i++ {
		dataPiece, err := os.ReadFile(dataPath + "/frame" + strconv.Itoa(i) + ".jpg")
		if err != nil {
			println("Error reading data: ", err.Error())
		}

		n, err := socket.WriteToUDP(dataPiece, addr)
		if err != nil {
			println("Error sending data: ", err.Error())
			continue
		}

		println("Sent data from client: ", n)
		time.Sleep(15 * time.Second)
	}
}

func receiveDataClient(socket *net.UDPConn, entityAddr string) {
	defer wg.Done()

	for {
		buffer := make([]byte, 65000)

		n, addr, err := socket.ReadFromUDP(buffer)
		if err != nil {
			println("Error reading from server: ", err.Error())
			continue
		}

		if addrStr := addr.String(); addrStr != entityAddr {
			dataBuffer := make([]byte, n)
			copy(dataBuffer, buffer[:n])
			println("Received data from entity at ", addrStr)
		}
	}
}

func receiveDataServer(socket *net.UDPConn, entityAddr string, canStream int) {
	defer wg.Done()

	for {
		buffer := make([]byte, 65000)

		n, addr, err := socket.ReadFromUDP(buffer)
		if err != nil {
			println("Error reading from server: ", err.Error())
			continue
		}

		if addrStr := addr.String(); addrStr != entityAddr {
			dataBuffer := make([]byte, n)
			copy(dataBuffer, buffer[:n])
			println("Received data from entity at ", addrStr)
			go sendAck(socket, addr)
		}
	}
}

func sendAck(socket *net.UDPConn, addr *net.UDPAddr) {
	_, err := socket.WriteToUDP([]byte("ACK"), addr)
	if err != nil {
		println("Error sending ACK: ", err.Error())
	} else {
		println("Sent ACK to ", addr.String())
	}
}

func getEntityAddr() string {
	hostname, err := os.Hostname()
	if err != nil {
		println("Error getting hostname: ", err.Error())
		os.Exit(0)
	}
	ips, err := net.LookupHost(hostname)
	if err != nil {
		println("Error resolving entity address: ", err.Error())
		os.Exit(0)
	}
	for _, ip := range ips {
		println("IP: ", ip)
	}
	entityAddr := ips[0] + ":" + os.Getenv("PORT")
	println("Entity address: ", entityAddr)
	return entityAddr
}

func applyMask(addr string) string {
	parts := strings.Split(addr, ".")
	parts[2] = "255"
	parts[3] = "255"
	return strings.Join(parts, ".")
}
