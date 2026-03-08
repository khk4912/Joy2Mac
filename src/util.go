package joy2mac

import (
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

func keepAlive(
	adapter *bluetooth.Adapter,
	deviceAddr bluetooth.Address) (bluetooth.Device, error) {

	for attempt := 1; attempt <= 3; attempt++ {

		device, err := adapter.Connect(deviceAddr, bluetooth.ConnectionParams{
			ConnectionTimeout: bluetooth.NewDuration(5 * time.Second),
		})
		if err == nil {
			fmt.Printf("Successfully reconnected to device %s on attempt %d\n", device.Address.UUID, attempt)
			return device, nil
		}

		fmt.Printf("Reconnection attempt %d failed: %v\n", attempt, err)
		time.Sleep(1 * time.Second)

	}

	return bluetooth.Device{}, fmt.Errorf("failed to reconnect after %d attempts", 3)
}
