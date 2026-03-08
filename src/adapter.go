package joy2mac

import (
	"sync"

	"tinygo.org/x/bluetooth"
)

type AdapterManager struct {
	adapter     *bluetooth.Adapter
	mu          sync.Mutex
	seenDevices map[string]struct{}
	candidates  []JoyconCandidate
}

func NewAdapterManager() *AdapterManager {
	return &AdapterManager{
		adapter:     bluetooth.DefaultAdapter,
		seenDevices: make(map[string]struct{}),
		candidates:  make([]JoyconCandidate, 0, maxFoundJoycons),
	}

}
