package main

import (
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {

	entity := os.Getenv("CONTAINER_NAME")
	ips, err := net.LookupHost(entity)
	if err != nil {
		println("Error resolving entity address: ", err.Error())
		os.Exit(0)
	}

	entityAddr := ips[0] + ":" + os.Getenv("PORT")
	println("Entity address: ", entityAddr)

	listenAddrRaw := "0.0.0.0" + ":" + os.Getenv("PORT")
	broadcastAddrRaw := "255.255.255.255" + ":" + os.Getenv("PORT")

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

	wg.Add(2)
	go sendData(socket, data, dataPath, broadcastAddr, canStream)
	go receiveData(socket, entityAddr)
	wg.Wait()
}

func sendData(socket *net.UDPConn, data []os.DirEntry, dataPath string, addr *net.UDPAddr, canStream int) {
	defer wg.Done()
	if canStream == 0 {
		return
	}
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
		time.Sleep(5 * time.Second)
	}
}

func receiveData(socket *net.UDPConn, entityAddr string) {
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
		} else {
			println("Filtered")
		}
	}
}
