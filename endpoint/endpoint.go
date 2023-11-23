package main

import (
	"net"
	"os"
	"strconv"
	"sync"
)

var wg sync.WaitGroup
var frameIndex int = 1

func main() {

	roleStr := os.Getenv("ROLE")

	entityAddr := getEntityAddr()
	entityAddrMasked := applyMask(entityAddr)

	listenAddrRaw := "0.0.0.0" + ":" + os.Getenv("PORT")
	broadcastAddrRaw := entityAddrMasked + ":" + os.Getenv("PORT")

	listenAddr := resolveAddr(listenAddrRaw)
	broadcastAddr := resolveAddr(broadcastAddrRaw)

	socket, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		println("Failed to establish broadcast connection: ", err.Error())
		os.Exit(0)
	}
	defer socket.Close()

	dataPath := os.Getenv("DATA_PATH")
	data, err := os.ReadDir(dataPath)
	if err != nil {
		println("Error reading data directory: ", err.Error())
		os.Exit(0)
	}

	switch roleStr {
	case "client":
		wg.Add(1)
		lookupEndpoint(socket, broadcastAddr)
		go receiveDataClient(socket, entityAddr, data, dataPath)
		wg.Wait()
	case "server":
		wg.Add(1)
		go receiveDataServer(socket, entityAddr)
		wg.Wait()
	default:
		println("Invalid role: ", roleStr)
	}
}

func lookupEndpoint(socket *net.UDPConn, addr *net.UDPAddr) {
	sourceID := prepID(os.Getenv("SOURCE_ID"))
	destID := prepID(os.Getenv("DEST_ID"))
	buffer := encode(make([]byte, 9), sourceID, 0, destID)

	_, err := socket.WriteToUDP(buffer, addr)
	if err != nil {
		println("Error sending data: ", err.Error())
	}
	println("Sent lookup")
}

func streamData(socket *net.UDPConn, data []os.DirEntry, dataPath string, addr *net.UDPAddr, sourceID, destID []int64) {
	dataPiece, err := os.ReadFile(dataPath + "/frame" + strconv.Itoa(frameIndex) + ".jpg")
	if err != nil {
		println("Error reading data: ", err.Error())
	}

	buffer := encode(make([]byte, 9+len(dataPiece)), sourceID, 1, destID)
	copy(buffer[9:], dataPiece)

	n, err := socket.WriteToUDP(buffer, addr)
	if err != nil {
		println("Error sending data: ", err.Error())
	}

	println("Sent data from client: ", n)
	frameIndex++
}

func receiveDataClient(socket *net.UDPConn, entityAddr string, data []os.DirEntry, dataPath string) {
	defer wg.Done()

	for {
		buffer := make([]byte, 65000)

		_, addr, err := socket.ReadFromUDP(buffer)
		if err != nil {
			println("Error reading from server: ", err.Error())
			continue
		}

		if addrStr := addr.String(); addrStr != entityAddr {
			source, transferType, dest := decodeToStr(buffer)
			if transferType == 2 {
				println("Received ACK from entity at ", addrStr)
				if frameIndex >= len(data) {
					sendInfo(socket, addr, encode(make([]byte, 9), prepID(dest), 3, prepID(source)))
				} else {
					go streamData(socket, data, dataPath, addr, prepID(dest), prepID(source))
				}
			}
		}
	}
}

func receiveDataServer(socket *net.UDPConn, entityAddr string) {
	defer wg.Done()

	for {
		buffer := make([]byte, 65000)

		_, addr, err := socket.ReadFromUDP(buffer)
		if err != nil {
			println("Error reading from server: ", err.Error())
			continue
		}

		source, transferType, dest := decodeToStr(buffer)

		if addrStr := addr.String(); addrStr != entityAddr && dest == os.Getenv("SOURCE_ID") {
			if transferType == 0 {
				println("Endpoint found!", addrStr)
				go sendInfo(socket, addr, encode(make([]byte, 9), prepID(dest), 2, prepID(source)))
			} else if transferType == 1 {
				println("Received data from ", source)
				go sendInfo(socket, addr, encode(make([]byte, 9), prepID(dest), 2, prepID(source)))
			}
		}
	}
}

func sendInfo(socket *net.UDPConn, addr *net.UDPAddr, buffer []byte) {
	_, err := socket.WriteToUDP(buffer, addr)
	if err != nil {
		println("Error sending ACK: ", err.Error())
	} else if buffer[4] == 2 {
		println("Sent ACK to ", addr.String())
	} else if buffer[4] == 3 {
		println("Sent REMOVE_REQUEST to ", addr.String())
	}
}
