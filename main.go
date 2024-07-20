package main

import (
	"fmt"
	"time"

	"syscall"
	"unsafe"

	"github.com/gen2brain/beeep"
	"github.com/go-vgo/robotgo"
	"github.com/tajtiattila/xinput"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetWindowRect       = user32.NewProc("GetWindowRect")
)

type RECT struct {
	left, top, right, bottom int32
}

func getForegroundWindow() (syscall.Handle, error) {
	ret, _, err := procGetForegroundWindow.Call()
	if ret == 0 {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

func getWindowRect(hwnd syscall.Handle) (*RECT, error) {
	var rect RECT
	ret, _, err := procGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rect)))
	if ret == 0 {
		return nil, err
	}
	return &rect, nil
}

func isFullScreen(rect *RECT) bool {
	// Get screen resolution
	screenWidth := int32(1920)  // Example resolution width, adjust as needed
	screenHeight := int32(1080) // Example resolution height, adjust as needed

	if rect.right-rect.left == screenWidth && rect.bottom-rect.top == screenHeight {
		return true
	}
	return false
}

func main() {

	isAppRunning := make(chan bool)
	fmt.Print("\033[H\033[2J")

	err := beeep.Notify("XNavigator", "XNavigator Running...", "assets/xbox.png")
	if err != nil {
		panic(err)
	}

	if err := xinput.Load(); err != nil {
		err := beeep.Notify("XNavigator", "Failed to load xinput", "assets/xbox.png")
		if err != nil {
			panic(err)
		}
		fmt.Println("Failed to load xinput:", err)
		fmt.Print("\033[H\033[2J")

		return
	}

	go checkFullScreenRunning(isAppRunning)
	var state xinput.State
	for {
		if err := xinput.GetState(0, &state); err != nil {
			fmt.Print("\033[H\033[2J")
			fmt.Println(err)
			time.Sleep(1000 * time.Millisecond)
		} else {
			fmt.Print("\033[H\033[2J")
			runner(&state, <-isAppRunning)
		}
	}

}

func checkFullScreenRunning(isRunning chan<- bool) {
	for {
		hwnd, err := getForegroundWindow()
		if err != nil {
			fmt.Print("\033[H\033[2J")

			fmt.Println("Error getting foreground window:", err)
			return
		}

		rect, err := getWindowRect(hwnd)
		if err != nil {
			fmt.Print("\033[H\033[2J")

			fmt.Println("Error getting window rect:", err)
			return
		}

		if isFullScreen(rect) {
			fmt.Print("\033[H\033[2J")
			fmt.Println("Fullscreen application detected!")
			panic("Fullscreen application detected!")
			isRunning <- true
		} else {
			fmt.Print("\033[H\033[2J")
			fmt.Println("No fullscreen application detected.")
			isRunning <- false
		}
		time.Sleep(1000 * time.Millisecond)
	}
}
func runner(st *xinput.State, isAppRunning bool) {
	if isAppRunning {
		time.Sleep(2000 * time.Millisecond)
	} else {
		runXNvigator(st)
	}
}

func runXNvigator(st *xinput.State) {
	state := *st
	err := beeep.Notify("XNavigator", "Controller found", "assets/xbox.png")
	fmt.Print("\033[H\033[2J")
	fmt.Println("Controller Detected")
	if err != nil {
		panic(err)
	}
	mousePositionX, mousePositiony := robotgo.Location()

	scalingFactor := 3000

	for {
		if err := xinput.GetState(0, &state); err != nil {
			err = beeep.Notify("XNavigator", "Controller disconnected", "assets/xbox.png")
			if err != nil {
				panic(err)
			}
			return
		}

		handleScale(&state, &scalingFactor)
		handleMouseMove(&state, &mousePositionX, &mousePositiony, &scalingFactor)
		handleMouseScroll(&state, &scalingFactor)

		time.Sleep(8 * time.Millisecond)

	}
}

func handleScale(state *xinput.State, scalingFactor *int) {
	if state.Gamepad.Buttons == uint16(xinput.LEFT_SHOULDER) {
		*scalingFactor = 700
	} else {
		*scalingFactor = 3000
	}
}

func handleMouseScroll(state *xinput.State, scalingFactor *int) {
	if state.Gamepad.ThumbRX != 0 || state.Gamepad.ThumbRY != 0 {
		robotgo.Scroll(int(state.Gamepad.ThumbRX) / *scalingFactor, int(state.Gamepad.ThumbRY)/(*scalingFactor*2))
	}

}

func handleMouseMove(state *xinput.State, mousePositionX *int, mousePositiony *int, sc *int) {

	scalingFactor := *sc

	*(mousePositionX) += int(state.Gamepad.ThumbLX) / scalingFactor
	*(mousePositiony) += int(state.Gamepad.ThumbLY) / scalingFactor * -1

	robotgo.Move(*(mousePositionX), *(mousePositiony))

	if state.Gamepad.Buttons == uint16(xinput.BUTTON_A) || state.Gamepad.Buttons == uint16(xinput.LEFT_THUMB) {
		robotgo.Click()
		time.Sleep(70 * time.Millisecond)
	}

	if state.Gamepad.Buttons == uint16(xinput.BUTTON_B) {
		robotgo.Click("right")
		time.Sleep(70 * time.Millisecond)
	}
}
