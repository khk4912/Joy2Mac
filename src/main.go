package joy2mac

import "fmt"

func StartSingleJoyconMode() {
	adapterManger := NewAdapterManager(1)
	candidates, _ := adapterManger.ScanJoycons()

	// TODO: err handling

	if len(candidates) != 1 {
		fmt.Println("Expected 1 Joy-Con device, found", len(candidates))
		fmt.Println("Stopping...")

		adapterManger.Shutdown()
		return
	}

	session := CreateJoyconSession(candidates[0], 1)
	adapterManger.AddJoyconSession(session)

	err := adapterManger.ConnectSession(session)
	if err != nil {
		fmt.Printf("Failed to connect to Joy-Con at %s: %v\n", candidates[0].AddressString, err)
		adapterManger.Shutdown()
		return
	}
}
