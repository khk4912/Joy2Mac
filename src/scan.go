package joy2mac

import (
	"fmt"
	"sync"
	"time"

	"tinygo.org/x/bluetooth"
)

const JOYCON_MANUFACTURER_ID = 0x553
const NINTENDO_SERVICE_UUID = "ab7de9be-89fe-49ad-828f-118f09df7fd0"

const INPUT_REPORT_CHARACTERISTIC_UUID = "ab7de9be-89fe-49ad-828f-118f09df7fd2"
const WRITE_COMMAND_CHARACTERISTIC_UUID = "649d4ac9-8eb7-4e6c-af44-1ea54fe5f005"

var JOYCON_MANUFACTURER_PREFIX = []byte{1, 0, 3, 126}

const (
	maxFoundJoycons = 2
	scanTimeout     = 10 * time.Second
)

type JoyconCandidate struct {
	Address       bluetooth.Address
	AddressString string
}

func ScanJoycons() []JoyconCandidate {
	adapter := bluetooth.DefaultAdapter

	err := adapter.Enable()
	if err != nil {
		panic(err)
	}

	var mu sync.Mutex
	var stopOnce sync.Once

	found := 0
	seen := map[string]struct{}{}
	candidates := make([]JoyconCandidate, 0, maxFoundJoycons)

	stopScan := func(a *bluetooth.Adapter, reason string) {
		stopOnce.Do(func() {
			fmt.Printf("Stopping scan: %s\n", reason)
			if stopErr := a.StopScan(); stopErr != nil {
				fmt.Printf("StopScan error: %v\n", stopErr)
			}
		})
	}

	timer := time.AfterFunc(scanTimeout, func() {
		stopScan(adapter, fmt.Sprintf("timeout (%s)", scanTimeout))
	})
	defer timer.Stop()

	fmt.Printf("Scanning for Joy-Con 2 (max %d, timeout %s)...\n\n", maxFoundJoycons, scanTimeout)

	err = adapter.Scan(func(a *bluetooth.Adapter, result bluetooth.ScanResult) {
		manufactureData := result.ManufacturerData()

		if len(manufactureData) == 0 {
			return
		}

		deviceInfo := manufactureData[0]
		if deviceInfo.CompanyID != JOYCON_MANUFACTURER_ID {
			return
		}

		if len(deviceInfo.Data) < len(JOYCON_MANUFACTURER_PREFIX) {
			return
		}

		for i, b := range JOYCON_MANUFACTURER_PREFIX {
			if deviceInfo.Data[i] != b {
				return
			}
		}

		addr := result.Address.String()

		mu.Lock()
		if _, exists := seen[addr]; exists {
			mu.Unlock()
			return
		}

		seen[addr] = struct{}{}

		found++
		count := found

		candidates = append(candidates, JoyconCandidate{
			Address:       result.Address,
			AddressString: addr,
		})

		mu.Unlock()

		fmt.Printf("Possible Joy-Con 2 found #%d\n", count)
		fmt.Printf("  Address: %s\n", addr)
		fmt.Printf("Try connecting to this BT device...\n")

		if count >= maxFoundJoycons {
			stopScan(a, fmt.Sprintf("found %d Joy-Con device(s)", maxFoundJoycons))
		}
	})

	if err != nil {
		panic(err)
	}

	fmt.Printf("Scan complete. Joy-Con candidates found: %d\n", found)
	return candidates
}

func StartJoyconConnection(candidate JoyconCandidate, playerNo int) (bluetooth.Device, error) {
	adapter := bluetooth.DefaultAdapter

	const maxAttempts = 3
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		fmt.Printf("\nAttempting to connect to device at %s (attempt %d/%d)...\n", candidate.AddressString, attempt, maxAttempts)

		device, err := adapter.Connect(candidate.Address, bluetooth.ConnectionParams{
			ConnectionTimeout: bluetooth.NewDuration(20 * time.Second),
		})
		if err != nil {
			lastErr = err
			fmt.Printf("Connect attempt failed: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err := onConnected(device, playerNo); err != nil {
			_ = device.Disconnect()
			return bluetooth.Device{}, fmt.Errorf("connected but setup failed: %w", err)
		}

		return device, nil
	}

	return bluetooth.Device{}, fmt.Errorf("failed to connect after %d attempts: %w", maxAttempts, lastErr)
}

func onConnected(device bluetooth.Device, playerNo int) error {
	fmt.Printf("Connected to device: %s\n", device.Address.UUID)

	nintendoServiceUUID, err := bluetooth.ParseUUID(NINTENDO_SERVICE_UUID)
	if err != nil {
		return fmt.Errorf("failed to parse Nintendo service UUID: %w", err)
	}

	writeUUID, err := bluetooth.ParseUUID(WRITE_COMMAND_CHARACTERISTIC_UUID)
	if err != nil {
		return fmt.Errorf("failed to parse write characteristic UUID: %w", err)
	}
	inputUUID, err := bluetooth.ParseUUID(INPUT_REPORT_CHARACTERISTIC_UUID)
	if err != nil {
		return fmt.Errorf("failed to parse input characteristic UUID: %w", err)
	}

	services, err := device.DiscoverServices([]bluetooth.UUID{nintendoServiceUUID})
	if err != nil {
		return fmt.Errorf("service discovery error: %w", err)
	}

	if len(services) == 0 {
		return fmt.Errorf("nintendo service not found on device %s", device.Address.UUID)
	}

	fmt.Printf("Services discovered: %d\n", len(services))
	z, err := services[0].DiscoverCharacteristics(nil)
	if err != nil {
		return fmt.Errorf("characteristic dump discovery failed: %w", err)
	}

	fmt.Println("Nintendo service found!")
	for _, c := range z {
		fmt.Printf("  Characteristic: %s\n", c.UUID())
	}

	nintendoService := services[0]
	writeCharacteristics, err := nintendoService.DiscoverCharacteristics([]bluetooth.UUID{writeUUID})
	if err != nil {
		return fmt.Errorf("write characteristic discovery error: %w", err)
	}
	if len(writeCharacteristics) == 0 {
		return fmt.Errorf("write characteristic not found on device %s", device.Address.UUID)
	}
	inputCharacteristics, err := nintendoService.DiscoverCharacteristics([]bluetooth.UUID{inputUUID})
	if err != nil {
		return fmt.Errorf("input characteristic discovery error: %w", err)
	}
	if len(inputCharacteristics) == 0 {
		return fmt.Errorf("input characteristic not found on device %s", device.Address.UUID)
	}

	fmt.Println(writeCharacteristics[0])

	fmt.Println("\nSetting player LEDs...")
	err = setPlayerLEDs(writeCharacteristics[0], playerNo)

	if err != nil {
		return fmt.Errorf("setPlayerLEDs failed: %w\n", err)
	}
	time.Sleep(150 * time.Millisecond)

	fmt.Println("Enabling IMU...")
	err = enable_imu(writeCharacteristics[0])
	if err != nil {
		return fmt.Errorf("enable_imu failed: %w\n", err)
	}

	fmt.Println("Enabling input notifications...")
	err = inputCharacteristics[0].EnableNotifications(func(buf []byte) {
		if len(buf) == 0 {
			return
		}
		fmt.Printf("Input report (%d): % X\n", len(buf), buf)
	})
	if err != nil {
		return fmt.Errorf("enable notifications failed: %w", err)
	}

	fmt.Println("Joy-Con notification stream is active.")
	return nil
}

func writeCommand(
	writeCharacteristic bluetooth.DeviceCharacteristic,
	commandID byte, subCommandID byte,
	cmd []byte) (int, error) {

	payload := make([]byte, len(cmd))
	copy(payload, cmd)

	if len(payload) < 8 {
		padding := make([]byte, 8-len(payload))
		payload = append(payload, padding...)
	}

	buffer := []byte{
		commandID,
		0x91,
		0x01,
		subCommandID,
		0x00,
		byte(len(payload)),
		0x00,
		0x00,
	}
	buffer = append(buffer, payload...)

	return writeCharacteristic.WriteWithoutResponse(buffer)
}

func setPlayerLEDs(writeCharacteristic bluetooth.DeviceCharacteristic, playerNo int) error {
	playerNo = max(1, min(playerNo, 8))

	LED_PATTERN_ID := map[int]byte{
		1: 0x01,
		2: 0x03,
		3: 0x07,
		4: 0x0F,
		5: 0x09,
		6: 0x05,
		7: 0x0D,
		8: 0x06,
	}

	// Send the LED command
	var LED_COMMAND_PREFIX byte = 0x09
	var SET_LED_COMMAND byte = 0x07

	_, err := writeCommand(writeCharacteristic, LED_COMMAND_PREFIX, SET_LED_COMMAND, []byte{LED_PATTERN_ID[playerNo]})
	if err != nil {
		return fmt.Errorf("failed to set player LEDs: %w", err)
	}

	return err
}

func enable_imu(writeCharacteristic bluetooth.DeviceCharacteristic) error {
	ENABLE_IMU_1 := []byte{0x0c, 0x91, 0x01, 0x02, 0x00, 0x04, 0x00, 0x00, 0x2f, 0x00, 0x00, 0x00}
	ENABLE_IMU_2 := []byte{0x0c, 0x91, 0x01, 0x04, 0x00, 0x04, 0x00, 0x00, 0x2f, 0x00, 0x00, 0x00}

	writeCharacteristic.WriteWithoutResponse(ENABLE_IMU_1)
	time.Sleep(500 * time.Millisecond)
	writeCharacteristic.WriteWithoutResponse(ENABLE_IMU_2)

	return nil
}
