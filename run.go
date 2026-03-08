package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	joy2mac "github.com/khk4912/joy2mac/src"
)

func main() {
	adapter := joy2mac.NewAdapterManager()
	candidates, err := adapter.ScanJoycons()

	var joyconSessions []*joy2mac.JoyconSession

	if err != nil {
		fmt.Printf("Error scanning for Joy-Con devices: %v\n", err)
		return
	}

	if len(candidates) == 0 {
		println("No Joy-Con candidates found. Exiting.")
		return
	}

	for i, candidate := range candidates {
		playerNo := i + 1
		session := joy2mac.CreateJoyconSession(candidate, playerNo)

		if err := session.StartJoyconConnection(); err != nil {
			fmt.Printf("Connection failed for %s: %v\n", candidate.AddressString, err)
		} else {
			joyconSessions = append(joyconSessions, session)
		}
	}

	if len(joyconSessions) == 0 {
		fmt.Println("No Joy-Con connection established, Exiting.")
		return
	}

	fmt.Println("Connected. Listening for input reports. Press Ctrl+C to exit.")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh

	fmt.Println("\nShutting down...")
	for _, session := range joyconSessions {
		_ = session.Device().Disconnect()
	}
}
