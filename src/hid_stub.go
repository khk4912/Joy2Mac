//go:build !darwin

// TODO:  darwin 아닐 시 구현 (언젠가...)
package joy2mac

type PointerButton uint8

const (
	PointerButtonLeft PointerButton = iota
	PointerButtonRight
)

func EnsurePostEventPermission(prompt bool) bool {
	return false
}

func MouseMove(dx, dy float64) {}

func MouseLeftButton(down bool) {}

func MouseRightButton(down bool) {}

func MouseDrag(dx, dy float64, button PointerButton) {}

func MouseScroll(delta int32) {}

func KeyToggle(code uint16, down bool) {}
