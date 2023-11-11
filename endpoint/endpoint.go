package main

import (
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

var wg sync.WaitGroup

var forwardingTable ForwardingTable

func main() {

	forwardingTable = ForwardingTable{
		entries: make(map[string]Hop),
	}

	sourceID := prepID(os.Getenv("SOURCE_ID"))
	destID := prepID(os.Getenv("DEST_ID"))
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
		lookupEndpoint(socket, broadcastAddr, sourceID, destID)
		go receiveDataClient(socket, entityAddr, data, dataPath)
		wg.Wait()
	case "server":
		wg.Add(1)
		go receiveDataServer(socket, entityAddr, sourceID)
		wg.Wait()
	default:
		println("Invalid role: ", roleStr)
	}
}

func lookupEndpoint(socket *net.UDPConn, addr *net.UDPAddr, sourceID, destID []int64) {
	buffer := encode(make([]byte, 9), sourceID, 0, destID)

	_, err := socket.WriteToUDP(buffer, addr)
	if err != nil {
		println("Error sending data: ", err.Error())
	}
}

func streamData(socket *net.UDPConn, data []os.DirEntry, dataPath string, addr *net.UDPAddr, sourceID, destID []int64) {
	defer wg.Done()
	for i := 1; i <= len(data); i++ {
		dataPiece, err := os.ReadFile(dataPath + "/frame" + strconv.Itoa(i) + ".jpg")
		if err != nil {
			println("Error reading data: ", err.Error())
		}

		buffer := encode(make([]byte, 9+len(dataPiece)), sourceID, 1, destID)
		copy(buffer[9:], dataPiece)

		n, err := socket.WriteToUDP(buffer, addr)
		if err != nil {
			println("Error sending data: ", err.Error())
			continue
		}

		println("Sent data from client: ", n)
		time.Sleep(5 * time.Second)
	}
}

func receiveDataClient(socket *net.UDPConn, entityAddr string, data []os.DirEntry, dataPath string) {
	defer wg.Done()
	var i int
	for {
		buffer := make([]byte, 65000)

		_, addr, err := socket.ReadFromUDP(buffer)
		if err != nil {
			println("Error reading from server: ", err.Error())
			continue
		}

		if addrStr := addr.String(); addrStr != entityAddr {
			println("Received ACK from entity at ", addrStr)
			source, transferType, dest := decodeToStr(buffer)
			if transferType == 2 {
				if _, exists := forwardingTable.GetRow(source); !exists {
					forwardingTable.AddRow(source, addr)
				}
				nextHop, _ := forwardingTable.GetRow(source)
				if i == 0 {
					go streamData(socket, data, dataPath, nextHop.IPAddress, prepID(dest), prepID(source))
				}
				i++
			}

		}
	}
}

func receiveDataServer(socket *net.UDPConn, entityAddr string, sourceID []int64) {
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
			if transferType == 0 {
				println("Endpoint found!", addrStr)
				forwardingTable.AddRow(source, addr)
			}
			go sendAck(socket, addr, encode(make([]byte, 9), prepID(dest), 2, prepID(source)))
		}
	}
}

func sendAck(socket *net.UDPConn, addr *net.UDPAddr, buffer []byte) {
	_, err := socket.WriteToUDP(buffer, addr)
	if err != nil {
		println("Error sending ACK: ", err.Error())
	} else {
		println("Sent ACK to ", addr.String())
	}
}
