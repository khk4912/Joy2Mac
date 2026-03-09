package joy2mac

import (
	"context"
)

func SingleJoyconHandler(ctx context.Context, inputCh <-chan InputData) {

	runJoyconTUI(ctx, []joyconInputSource{
		{name: "P1", title: "Single Joy-Con", inputCh: inputCh, side: UnknownSide},
	})
}

func DualJoyconHandler(ctx context.Context, leftInputCh <-chan InputData, rightInputCh <-chan InputData) {
	runJoyconTUI(ctx, []joyconInputSource{
		{name: "P1", title: "Left Joy-Con 2", inputCh: leftInputCh, side: LeftSide},
		{name: "P2", title: "Right Joy-Con 2", inputCh: rightInputCh, side: RightSide},
	})
}
