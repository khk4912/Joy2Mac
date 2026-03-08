package joy2mac

import (
	"errors"
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

const NINTENDO_SERVICE_UUID = "ab7de9be-89fe-49ad-828f-118f09df7fd0"
const INPUT_REPORT_CHARACTERISTIC_UUID = "ab7de9be-89fe-49ad-828f-118f09df7fd2"
const WRITE_COMMAND_CHARACTERISTIC_UUID = "649d4ac9-8eb7-4e6c-af44-1ea54fe5f005"

var ErrNintendoServiceNotFound = errors.New("nintendo service not found")

type JoyconSession struct {
	address             bluetooth.Address
	device              bluetooth.Device
	nintendoService     bluetooth.DeviceService
	writeCharacteristic bluetooth.DeviceCharacteristic
	inputCharacteristic bluetooth.DeviceCharacteristic
	playerNo            int

	// State flags
	Connected                bool
	writeCharacteristicFound bool
	inputCharacteristicFound bool
}

func CreateJoyconSession(
	candidate JoyconCandidate,
	playerNo int,
) *JoyconSession {
	return &JoyconSession{
		address:  candidate.Address,
		playerNo: playerNo,
	}
}

func (session *JoyconSession) Device() bluetooth.Device {
	return session.device
}

func (session *JoyconSession) StartJoyconConnection() error {
	device, err := bluetooth.DefaultAdapter.Connect(session.address, bluetooth.ConnectionParams{
		ConnectionTimeout: bluetooth.NewDuration(5 * time.Second),
	})

	if err != nil {
		return err
	}

	session.device = device
	if err := session.setupServices(); err != nil {
		_ = device.Disconnect()
		return err
	}

	session.Connected = true
	return nil

}

func (session *JoyconSession) setupServices() error {
	nintendoServiceUUID, err := bluetooth.ParseUUID(NINTENDO_SERVICE_UUID)
	if err != nil {
		return err
	}

	services, err := session.discoverService(nintendoServiceUUID)
	if err != nil {
		return fmt.Errorf("failed to discover nintendo service: %w", err)
	}

	if len(services) == 0 {
		return ErrNintendoServiceNotFound
	}

	session.nintendoService = services[0]

	writeUUID, err := bluetooth.ParseUUID(WRITE_COMMAND_CHARACTERISTIC_UUID)
	if err != nil {
		return fmt.Errorf("failed to parse write characteristic UUID: %w", err)
	}

	inputUUID, err := bluetooth.ParseUUID(INPUT_REPORT_CHARACTERISTIC_UUID)
	if err != nil {
		return fmt.Errorf("failed to parse input characteristic UUID: %w", err)
	}

	session.writeCharacteristic, err = session.discoverCharacteristic(writeUUID)
	writeChar, err := session.discoverCharacteristic(writeUUID)
	if err != nil {
		return fmt.Errorf("failed to discover write characteristic: %w", err)
	}

	inputChar, err := session.discoverCharacteristic(inputUUID)
	if err != nil {
		return fmt.Errorf("failed to discover input characteristic: %w", err)
	}

	session.writeCharacteristic = writeChar
	session.inputCharacteristic = inputChar

	return nil
}

func (session *JoyconSession) discoverService(serviceUUID bluetooth.UUID) ([]bluetooth.DeviceService, error) {
	services, err := session.device.DiscoverServices([]bluetooth.UUID{serviceUUID})

	if err != nil {
		return nil, fmt.Errorf("service discovery error: %w", err)
	}
	return services, nil
}

func (session *JoyconSession) discoverCharacteristic(characteristicUUID bluetooth.UUID) (bluetooth.DeviceCharacteristic, error) {
	characteristics, err := session.nintendoService.DiscoverCharacteristics([]bluetooth.UUID{characteristicUUID})
	if err != nil {
		return bluetooth.DeviceCharacteristic{}, fmt.Errorf("characteristic discovery error: %w", err)
	}

	return characteristics[0], nil
}

func (session *JoyconSession) writeCommand(
	commandID byte,
	subCommandID byte,
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

	return session.writeCharacteristic.WriteWithoutResponse(buffer)
}

func (session *JoyconSession) setPlayerLEDs(playerNo int) error {
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

	_, err := session.writeCommand(LED_COMMAND_PREFIX, SET_LED_COMMAND, []byte{LED_PATTERN_ID[playerNo]})
	if err != nil {
		return fmt.Errorf("failed to set player LEDs: %w", err)
	}

	return err
}

func (session *JoyconSession) enable_imu() error {
	ENABLE_IMU_1 := []byte{0x0c, 0x91, 0x01, 0x02, 0x00, 0x04, 0x00, 0x00, 0x2f, 0x00, 0x00, 0x00}
	ENABLE_IMU_2 := []byte{0x0c, 0x91, 0x01, 0x04, 0x00, 0x04, 0x00, 0x00, 0x2f, 0x00, 0x00, 0x00}

	session.writeCharacteristic.WriteWithoutResponse(ENABLE_IMU_1)
	time.Sleep(500 * time.Millisecond)
	session.writeCharacteristic.WriteWithoutResponse(ENABLE_IMU_2)

	return nil
}
