package trade

import (
	"black_hole/data"
	"testing"
	"time"
)

func TestAlarm(t *testing.T) {
	zbTradeAPI := ZbLive{
		accessKey:  "7e9c8d57-d23a-4ee6-a99f-15f72b305851",
		sercetKey:  "104eb581-bd14-4f03-a8a2-a48cd72a3a4c",
		focusCoins: []string{"QC", "ETH", "BTC"},
	}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	for {
		BroadAsset(&zbTradeAPI)
		time.Sleep(time.Second)
	}
}
