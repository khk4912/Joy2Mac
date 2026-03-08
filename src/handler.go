package joy2mac

import (
	"context"
	"fmt"
)

func SingleJoyconHandler(ctx context.Context, inputCh <-chan InputData) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Joy-Con handler stopped.")
			return
		case input := <-inputCh:
			fmt.Printf("Received input: %+v\n", input)
		default:
		}

	}
}
