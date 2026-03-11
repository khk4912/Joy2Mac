//go:build darwin

package joy2mac

/*
#cgo darwin LDFLAGS: -framework ApplicationServices
#include <ApplicationServices/ApplicationServices.h>
#include <stdbool.h>
#include <stdint.h>

static bool joy_preflight_post_access() {
	return CGPreflightPostEventAccess();
}

static bool joy_request_post_access() {
	return CGRequestPostEventAccess();
}

static CGPoint joy_current_mouse_pos() {
	CGEventRef event = CGEventCreate(NULL);
	CGPoint p = CGEventGetLocation(event);
	CFRelease(event);
	return p;
}

static void joy_post_mouse_event(CGEventType typ, CGMouseButton button, double x, double y) {
	CGEventRef event = CGEventCreateMouseEvent(NULL, typ, CGPointMake(x, y), button);
	CGEventPost(kCGHIDEventTap, event);
	CFRelease(event);
}

static void joy_post_mouse_move(double x, double y) {
	joy_post_mouse_event(kCGEventMouseMoved, kCGMouseButtonLeft, x, y);
}

static void joy_post_left_mouse(bool down, double x, double y) {
	CGEventType typ = down ? kCGEventLeftMouseDown : kCGEventLeftMouseUp;
	joy_post_mouse_event(typ, kCGMouseButtonLeft, x, y);
}

static void joy_post_right_mouse(bool down, double x, double y) {
	CGEventType typ = down ? kCGEventRightMouseDown : kCGEventRightMouseUp;
	joy_post_mouse_event(typ, kCGMouseButtonRight, x, y);
}

static void joy_post_mouse_drag(double x, double y, uint8_t button) {
	CGEventType typ = button == 1 ? kCGEventRightMouseDragged : kCGEventLeftMouseDragged;
	CGMouseButton mouseButton = button == 1 ? kCGMouseButtonRight : kCGMouseButtonLeft;
	joy_post_mouse_event(typ, mouseButton, x, y);
}

static void joy_post_scroll(int32_t dy) {
	CGEventRef event = CGEventCreateScrollWheelEvent(NULL, kCGScrollEventUnitLine, 1, dy);
	CGEventPost(kCGHIDEventTap, event);
	CFRelease(event);
}

static void joy_post_key(uint16_t keycode, bool down) {
	CGEventRef event = CGEventCreateKeyboardEvent(NULL, (CGKeyCode)keycode, down);
	CGEventPost(kCGHIDEventTap, event);
	CFRelease(event);
}
*/
import "C"

type PointerButton uint8

const (
	PointerButtonLeft PointerButton = iota
	PointerButtonRight
)

func EnsurePostEventPermission(prompt bool) bool {
	if bool(C.joy_preflight_post_access()) {
		return true
	}
	if prompt {
		return bool(C.joy_request_post_access())
	}
	return false
}

func MouseMove(dx, dy float64) {
	pos := C.joy_current_mouse_pos()
	C.joy_post_mouse_move(pos.x+C.double(dx), pos.y+C.double(dy))
}

func MouseLeftButton(down bool) {
	pos := C.joy_current_mouse_pos()
	C.joy_post_left_mouse(C.bool(down), pos.x, pos.y)
}

func MouseRightButton(down bool) {
	pos := C.joy_current_mouse_pos()
	C.joy_post_right_mouse(C.bool(down), pos.x, pos.y)
}

func MouseDrag(dx, dy float64, button PointerButton) {
	pos := C.joy_current_mouse_pos()
	C.joy_post_mouse_drag(pos.x+C.double(dx), pos.y+C.double(dy), C.uint8_t(button))
}

func MouseScroll(delta int32) {
	C.joy_post_scroll(C.int32_t(delta))
}

func KeyToggle(code uint16, down bool) {
	C.joy_post_key(C.uint16_t(code), C.bool(down))
}
