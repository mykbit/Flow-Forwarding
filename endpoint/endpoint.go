package main

import (
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	TransferTypeBroadcast = iota
	TransferTypeData
	TransferTypeAck
	TransferTypeRemove
)

var wg sync.WaitGroup
var frameIndex int = 1

func main() {
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

	wg.Add(1)
	go receiveData(socket, entityAddr, data, dataPath, broadcastAddr)
	wg.Wait()
}

func lookupEndpoint(socket *net.UDPConn, addr *net.UDPAddr) {
	time.Sleep(1 * time.Second)

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

func receiveData(socket *net.UDPConn, entityAddr string, data []os.DirEntry, dataPath string, broadcastAddr *net.UDPAddr) {
	defer wg.Done()

	go lookupEndpoint(socket, broadcastAddr)

	for {
		buffer := make([]byte, 65000)

		_, addr, err := socket.ReadFromUDP(buffer)
		if err != nil {
			println("Error reading from server: ", err.Error())
			continue
		}

		source, transferType, dest := decodeToStr(buffer)
		if addrStr := addr.String(); addrStr != entityAddr && dest == os.Getenv("SOURCE_ID") {
			switch transferType {
			case TransferTypeBroadcast:
				println("Endpoint found!", addrStr)
				go sendInfo(socket, addr, encode(make([]byte, 9), prepID(dest), 2, prepID(source)))
			case TransferTypeData:
				println("Received data from ", source)
				go sendInfo(socket, addr, encode(make([]byte, 9), prepID(dest), 2, prepID(source)))
			case TransferTypeAck:
				println("Received ACK from entity at ", addrStr)
				if frameIndex >= len(data) {
					sendInfo(socket, addr, encode(make([]byte, 9), prepID(dest), 3, prepID(source)))
					break
				} else {
					go streamData(socket, data, dataPath, addr, prepID(dest), prepID(source))
				}
			}
		}
	}
}

func sendInfo(socket *net.UDPConn, addr *net.UDPAddr, buffer []byte) {
	_, err := socket.WriteToUDP(buffer, addr)
	if err != nil {
		println("Error sending ACK: ", err.Error())
		return
	} else if buffer[4] == 2 {
		println("Sent ACK to ", addr.String())
	} else if buffer[4] == 3 {
		println("Sent REMOVE_REQUEST to ", addr.String())
	}
}
