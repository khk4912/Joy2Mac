package joy2mac

import (
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

const JOYCON_MANUFACTURER_ID = 0x553
const SCAN_TIMEOUT = 10 * time.Second

var JOYCON_MANUFACTURER_PREFIX = []byte{1, 0, 3, 126}

type JoyconCandidate struct {
	Device        bluetooth.Device
	Address       bluetooth.Address
	AddressString string
}

func (am *AdapterManager) ScanJoycons() ([]JoyconCandidate, error) {
	adapter := am.Adapter

	err := adapter.Enable()
	if err != nil {
		return nil, err
	}

	timer := time.AfterFunc(SCAN_TIMEOUT, func() {
		if err :=
			am.stopScan(fmt.Sprintf("timeout (%s)", SCAN_TIMEOUT)); err != nil {
		}
	})
	defer timer.Stop()

	fmt.Printf("Scanning for Joy-Con 2 (max %d, timeout %s)...\n\n", am.maxJoyconConnections, SCAN_TIMEOUT)

	err = adapter.Scan(
		func(a *bluetooth.Adapter, result bluetooth.ScanResult) {
			err := am.onAdapterScan(result)

			if err != nil {
				fmt.Printf("Error during scan callback: %v\n", err)
			}
		})

	if err != nil {
		return nil, fmt.Errorf("Failed to scan for Joy-Con devices: %w", err)
	}

	fmt.Printf("Scan complete. Joy-Con candidates found: %d\n", len(am.candidates))
	return am.candidates, nil
}

func (am *AdapterManager) stopScan(reason string) error {
	fmt.Printf("Stopping scan: %s\n", reason)

	if stopErr := am.Adapter.StopScan(); stopErr != nil {
		return fmt.Errorf("Failed to stop scan: %w", stopErr)
	}

	return nil
}

func (am *AdapterManager) onAdapterScan(result bluetooth.ScanResult) error {
	manufactureData := result.ManufacturerData()

	if len(manufactureData) == 0 {
		return nil
	}

	deviceInfo := manufactureData[0]
	if deviceInfo.CompanyID != JOYCON_MANUFACTURER_ID {
		return nil
	}

	if len(deviceInfo.Data) < len(JOYCON_MANUFACTURER_PREFIX) {
		return nil
	}

	for i, b := range JOYCON_MANUFACTURER_PREFIX {
		if deviceInfo.Data[i] != b {
			return nil
		}
	}

	addr := result.Address.String()

	if _, exists := am.seenDevices[addr]; exists {
		return nil
	}

	am.mu.Lock()
	am.seenDevices[addr] = struct{}{}
	am.candidates = append(am.candidates, JoyconCandidate{
		Address:       result.Address,
		AddressString: addr,
	})
	am.mu.Unlock()

	fmt.Printf("Possible Joy-Con 2 found #%d\n", len(am.candidates))
	fmt.Printf("  Address: %s\n", addr)

	if len(am.candidates) >= am.maxJoyconConnections {
		if err := am.stopScan(fmt.Sprintf("found %d Joy-Con device(s)", am.maxJoyconConnections)); err != nil {
			return fmt.Errorf("Error stopping scan: %w", err)
		}
	}

	return nil
}
