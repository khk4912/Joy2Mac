package joy2mac

import (
	"encoding/binary"
)

func parseInputData(joyconInputch <-chan InputData) <-chan JoyconState {
	parsedCh := make(chan JoyconState, 1)

	go func() {
		defer close(parsedCh)

		for input := range joyconInputch {
			raw := append([]byte(nil), input.Data...)
			state := JoyconState{
				PlayerNo:    input.PlayerNo,
				Side:        input.Side,
				Raw:         raw,
				Buttons:     ButtonState{},
				Temperature: parseTemperature(raw),
				Accel:       parseAccel(raw),
				Gyro:        parseGyro(raw),
				Voltage:     parseVoltage(raw),
			}

			switch state.Side {
			case LeftSide:
				state.Stick = parseLeftStick(raw)
				state.Buttons = parseLeftButtons(raw)
			case RightSide:
				// state.Buttons = parseRightButtons(raw)
				// state.Stick = parseRightStick(raw)
			}

			parsedCh <- state
		}
	}()

	return parsedCh
}

func read(data []byte, offset int, length int) []byte {
	if offset+length > len(data) {
		return nil
	}
	return data[offset : offset+length]
}

func parseTemperature(data []byte) float64 {
	raw := read(data, 0x2E, 2)
	if raw == nil {
		return 0
	}
	tempRaw := binary.LittleEndian.Uint16(raw)

	// 25°C + raw / 127
	return 25 + float64(int16(tempRaw))/127
}

func parseVoltage(data []byte) float64 {
	raw := read(data, 0x1C, 2)
	if raw == nil {
		return 0
	}
	voltageRaw := binary.LittleEndian.Uint16(raw)

	return float64(voltageRaw) * 0.001
}

func parseAccel(data []byte) [3]float64 {
	raw := read(data, 0x30, 6)
	if raw == nil {
		return [3]float64{}
	}

	Xraw := binary.LittleEndian.Uint16(raw[0:2])
	Yraw := binary.LittleEndian.Uint16(raw[2:4])
	Zraw := binary.LittleEndian.Uint16(raw[4:6])

	// 4096 equals 1G, returns acceleration in G
	return [3]float64{
		float64(int16(Xraw)) / 4096,
		float64(int16(Yraw)) / 4096,
		float64(int16(Zraw)) / 4096,
	}
}

func parseGyro(data []byte) [3]float64 {
	// 48000 = 360°/s, returns angular velocity in °/s

	raw := read(data, 0x36, 6)
	if raw == nil {
		return [3]float64{}
	}

	Xraw := binary.LittleEndian.Uint16(raw[0:2])
	Yraw := binary.LittleEndian.Uint16(raw[2:4])
	Zraw := binary.LittleEndian.Uint16(raw[4:6])

	return [3]float64{
		float64(int16(Xraw)) / 48000 * 360,
		float64(int16(Yraw)) / 48000 * 360,
		float64(int16(Zraw)) / 48000 * 360,
	}
}

func parseLeftButtons(data []byte) ButtonState {
	raw := read(data, 0x04, 3)
	if raw == nil {
		return ButtonState{}
	}
	// fmt.Printf("buttons raw: %02x %02x %02x %02x\n", raw[0], raw[1], raw[2], raw[3])
	return ButtonState{
		ButtonDown:    raw[2]&0x01 != 0,
		ButtonUp:      raw[2]&0x02 != 0,
		ButtonRight:   raw[2]&0x04 != 0,
		ButtonLeft:    raw[2]&0x08 != 0,
		ButtonSR:      raw[2]&0x10 != 0,
		ButtonSL:      raw[2]&0x20 != 0,
		ButtonL:       raw[2]&0x40 != 0,
		ButtonZL:      raw[2]&0x80 != 0,
		ButtonMinus:   raw[1]&0x01 != 0,
		ButtonStick:   raw[1]&0x08 != 0,
		ButtonCapture: raw[1]&0x20 != 0,
	}
}

func parseLeftStick(data []byte) StickInput {
	raw := read(data, 10, 3)
	if raw == nil {
		return StickInput{}
	}

	rawNum := binary.LittleEndian.Uint32(append(raw, 0x00)) // Pad to 4 bytes for easier bit manipulation

	return StickInput{
		X: int16(rawNum & 0xFFF),
		Y: int16((rawNum >> 12) & 0xFFF),
	}
}
