package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

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

func prepID(id string) []int64 {
	parts := strings.Split(id, ":")
	if parts[0] == "" {
		return nil
	}
	var idParts []int64
	for _, part := range parts {
		id_slice, err := strconv.ParseInt(part, 16, 64)
		if err != nil {
			println("Error parsing ID: ", err.Error())
			os.Exit(0)
		}
		idParts = append(idParts, id_slice)
	}
	return idParts
}

func encode(buffer []byte, sourceID []int64, transferType int, destID []int64) []byte {
	buffer[0] = byte(sourceID[0])
	buffer[1] = byte(sourceID[1])
	buffer[2] = byte(sourceID[2])
	buffer[3] = byte(sourceID[3])
	buffer[4] = byte(transferType)
	buffer[5] = byte(destID[0])
	buffer[6] = byte(destID[1])
	buffer[7] = byte(destID[2])
	buffer[8] = byte(destID[3])

	return buffer
}

func decodeToInt(buffer []byte) ([]int64, int, []int64) {
	var sourceID []int64
	var transferType int
	var destID []int64

	sourceID = append(sourceID, int64(buffer[0]))
	sourceID = append(sourceID, int64(buffer[1]))
	sourceID = append(sourceID, int64(buffer[2]))
	sourceID = append(sourceID, int64(buffer[3]))
	transferType = int(buffer[4])
	destID = append(destID, int64(buffer[5]))
	destID = append(destID, int64(buffer[6]))
	destID = append(destID, int64(buffer[7]))
	destID = append(destID, int64(buffer[8]))

	return sourceID, transferType, destID

}

func decodeToStr(buffer []byte) (string, int, string) {
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

func resolveAddr(addr string) *net.UDPAddr {
	resolvedAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		println("Failed to resolve address: ", err.Error())
		os.Exit(0)
	}
	return resolvedAddr
}
