package joy2mac

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func StartSingleJoyconMode() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	adapterManger := NewAdapterManager(1)
	candidates, err := adapterManger.ScanJoycons()

	if err != nil {
		fmt.Printf("Failed to scan Joy-Con devices: %v\n", err)
		adapterManger.Shutdown()
		return
	}

	if len(candidates) != 1 {
		fmt.Println("Expected 1 Joy-Con device, found", len(candidates))
		fmt.Println("Stopping...")

		adapterManger.Shutdown()
		return
	}

	inputCh := make(chan InputData, 1)
	session := CreateJoyconSession(candidates[0], 1, inputCh)
	adapterManger.AddJoyconSession(session)

	err = adapterManger.ConnectSession(session)
	if err != nil {
		fmt.Printf("Failed to connect to Joy-Con at %s: %v\n", candidates[0].AddressString, err)
		adapterManger.Shutdown()
		return
	}

	go SingleJoyconHandler(ctx, inputCh)
	<-ctx.Done()

	fmt.Println("\nShutting down...")
	adapterManger.Shutdown()

}
