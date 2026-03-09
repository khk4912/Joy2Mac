package joy2mac

type JoyconSide int

const (
	UnknownSide JoyconSide = iota
	LeftSide
	RightSide
)

type Button int

const (
	ButtonUnknown Button = iota
	ButtonUp
	ButtonDown
	ButtonLeft
	ButtonRight
	ButtonA
	ButtonB
	ButtonX
	ButtonY
	ButtonL
	ButtonR
	ButtonZL
	ButtonZR
	ButtonSL
	ButtonSR
	ButtonPlus
	ButtonMinus
	ButtonHome
	ButtonCapture
	ButtonStick
	ButtonGameChat
)

type StickInput struct {
	X float64
	Y float64
}

type InputData struct {
	PlayerNo int
	Side     JoyconSide
	Data     []byte
}

type ButtonState map[Button]bool

type JoyconState struct {
	PlayerNo int
	Side     JoyconSide
	Stick    StickInput
	Buttons  ButtonState
}

var State JoyconState
