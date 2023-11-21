package main

import (
	"net"
	"sync"
	"time"
)

type Hop struct {
	IPAddress *net.UDPAddr
	ExpireCh  chan bool
}

type ForwardingTable struct {
	mu      sync.Mutex
	entries map[string]Hop
}

func (f *ForwardingTable) AddRow(destinationID string, ipAddress *net.UDPAddr) {
	f.mu.Lock()
	defer f.mu.Unlock()

	expireCh := make(chan bool)

	hop := Hop{
		IPAddress: ipAddress,
		ExpireCh:  expireCh,
	}

	f.entries[destinationID] = hop

	go func() {
		select {
		case <-time.After(5 * time.Minute):
			f.RemoveRow(destinationID)
		case <-expireCh:
			// Do nothing, the destination was removed explicitly
		}
	}()
}

func (f *ForwardingTable) RemoveRow(destinationID string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if hop, exists := f.entries[destinationID]; exists {
		close(hop.ExpireCh)
		delete(f.entries, destinationID)
		println("Removed destination: ", destinationID)
	}
}

func (f *ForwardingTable) GetRow(destinationID string) (Hop, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	hop, exists := f.entries[destinationID]
	return hop, exists
}
