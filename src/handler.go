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
		case input, ok := <-inputCh:
			if !ok {
				fmt.Println("Input channel closed.")
				return
			}
			fmt.Printf("Received input: %+v\n", input)
		}

	}
}

func DualJoyconHandler(ctx context.Context, leftInputCh <-chan InputData, rightInputCh <-chan InputData) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Joy-Con handler stopped.")
			return
		case input, ok := <-leftInputCh:
			if !ok {
				fmt.Println("Left Joy-Con input channel closed.")
				return
			}
			fmt.Printf("Received input from left Joy-Con: %+v\n", input)
		case input, ok := <-rightInputCh:
			if !ok {
				fmt.Println("Right Joy-Con input channel closed.")
				return
			}
			fmt.Printf("Received input from right Joy-Con: %+v\n", input)
		}

	}
}
