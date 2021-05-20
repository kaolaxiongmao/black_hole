package trade

import (
	"black_hole/constdef"
	"black_hole/dal"
	"black_hole/data"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestZbLive_GetAccountInfo(t *testing.T) {
	zbTradeAPI := NewTradeAPI()
	zbTradeAPI.crossTradePairs =  []string{"btcqc", "ethqc", "ltcqc"}
	zbTradeAPI.crossTradePairs =  []string{"btcqc", "ethqc", "ltcqc"}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	zbTradeAPI.GetAccountInfo()
	time.Sleep(time.Second)
}

func TestZbLive_SetOrderBuy(t *testing.T) {
	zbTradeAPI := ZbLive{
		accessKey:  "7e9c8d57-d23a-4ee6-a99f-15f72b305851",
		sercetKey:  "104eb581-bd14-4f03-a8a2-a48cd72a3a4c",
		focusCoins: []string{"QC", "ETH"},
	}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	zbTradeAPI.SetOrder(int(constdef.Buy), "BTC", 500, 1)
	time.Sleep(time.Second * 3)
}

func TestZbLive_SetOrderSell(t *testing.T) {
	zbTradeAPI := ZbLive{
		accessKey:  "7e9c8d57-d23a-4ee6-a99f-15f72b305851",
		sercetKey:  "104eb581-bd14-4f03-a8a2-a48cd72a3a4c",
		focusCoins: []string{"QC", "ETH"},
	}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	time.Sleep(time.Second * 10)
}

func TestZbLive_GetCrossAsset(t *testing.T) {
	zbTradeAPI := NewTradeAPI()
	zbTradeAPI.crossTradePairs =  []string{"btcqc"}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	zbTradeAPI.GetCrossAsset()
	time.Sleep(time.Second)
}

func TestZbLive_DoCrossLoan(t *testing.T) {
	zbTradeAPI := ZbLive{
		accessKey:  "7e9c8d57-d23a-4ee6-a99f-15f72b305851",
		sercetKey:  "104eb581-bd14-4f03-a8a2-a48cd72a3a4c",
		passWord: "053641",
		focusCoins: []string{"QC", "ETH", "BTC"},
	}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	zbTradeAPI.DoCrossLoan(200, "qc", "btcqc")
}
func TestZbLive_DoCrossRepay(t *testing.T) {
	zbTradeAPI := ZbLive{
		accessKey:  "7e9c8d57-d23a-4ee6-a99f-15f72b305851",
		sercetKey:  "104eb581-bd14-4f03-a8a2-a48cd72a3a4c",
		focusCoins: []string{"QC", "ETH", "BTC"},
	}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	zbTradeAPI.DoCrossRepay("qc", "btcqc")
}

func TestZbLive_CancelOrder(t *testing.T) {
	zbTradeAPI := ZbLive{
		accessKey:  "7e9c8d57-d23a-4ee6-a99f-15f72b305851",
		sercetKey:  "104eb581-bd14-4f03-a8a2-a48cd72a3a4c",
		focusCoins: []string{"QC", "ETH", "BTC"},
	}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	id, _, _, err := zbTradeAPI.SetOrder(int(constdef.Buy), "BTC", 200, 1)
	if err != nil {
		log.Fatalf("SetOrder： err: %v", err)
		return
	}
	fmt.Println("id ", id, " err ", err)
	zbTradeAPI.QueryOrderStatus("btc_qc", id)
	zbTradeAPI.CancelOrder("btc_qc", id)
	time.Sleep(time.Second)
}
func TestZbLive_CancelOrder2(t *testing.T) {
	zbTradeAPI := ZbLive{
		accessKey:  "7e9c8d57-d23a-4ee6-a99f-15f72b305851",
		sercetKey:  "104eb581-bd14-4f03-a8a2-a48cd72a3a4c",
		focusCoins: []string{"QC", "ETH", "BTC"},
	}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	err := zbTradeAPI.CancelOrder("btc_qc", "202012282799782258")
	if err != nil {
		log.Fatalf("CacelOrder: %v", err)
	}
	time.Sleep(time.Second)
}
func TestZbLive_QueryOrder(t *testing.T) {
	zbTradeAPI := ZbLive{
		accessKey:  "7e9c8d57-d23a-4ee6-a99f-15f72b305851",
		sercetKey:  "104eb581-bd14-4f03-a8a2-a48cd72a3a4c",
		focusCoins: []string{"QC", "ETH", "BTC"},
	}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	zbTradeAPI.QueryOrderStatus("btc_qc", "202012282799942186")
	time.Sleep(time.Second)
}

func TestZbLive_OpenMutipleOrder(t *testing.T) {
	dal.OrderStorage.Del([]string{constdef.LatestSingalKey, constdef.LatestSingalStatusKey}...)
	singalKey := "singalKey_test_1"
	zbTradeAPI := NewTradeAPI()
	zbTradeAPI.crossTradePairs =  []string{"btcqc", "ethqc", "ltcqc"}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	for {
		OpenMutipleOrder(zbTradeAPI, singalKey, "BTC")
		time.Sleep(time.Second*2)
	}
}
func TestZbLive_CloseMutipleOrder(t *testing.T) {
	zbTradeAPI := NewTradeAPI()
	zbTradeAPI.crossTradePairs =  []string{"btcqc", "ethqc", "ltcqc"}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	for {
		CloseMutipleOrder(zbTradeAPI,  "BTC", "singalKey_test_1")
		time.Sleep(time.Second*2)
	}
}
func TestZbLive_OpenEmptyOrder(t *testing.T) {
	dal.OrderStorage.Del([]string{constdef.LatestSingalKey, constdef.LatestSingalStatusKey}...)
	singalKey := "singalKey_test_2"
	zbTradeAPI := NewTradeAPI()
	zbTradeAPI.crossTradePairs =  []string{"btcqc", "ethqc", "ltcqc"}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	for {
		OpenEmptyOrder(zbTradeAPI, singalKey, "BTC")
		time.Sleep(time.Second*2)
	}
}
func TestZbLive_CloseEmptyOrder(t *testing.T) {
	zbTradeAPI := NewTradeAPI()
	zbTradeAPI.crossTradePairs =  []string{"btcqc", "ethqc", "ltcqc"}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	for {
		CloseEmptyOrder(zbTradeAPI,  "BTC", "singalKey_test_2")
		time.Sleep(time.Second*2)
	}
}

func  TestZbLive_BuySingal(t *testing.T) {
	zbTradeAPI := NewTradeAPI()
	zbTradeAPI.crossTradePairs =  []string{"btcqc", "ethqc", "ltcqc"}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	for {
		OpenOrder(zbTradeAPI,  constdef.Buy, "singal_test_52", "BTC")
		time.Sleep(time.Second*2)
	}
}

func TestZbLive_SellSingal(t *testing.T) {
	zbTradeAPI := NewTradeAPI()
	zbTradeAPI.crossTradePairs =  []string{"btcqc", "ethqc", "ltcqc"}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	for {
		OpenOrder(zbTradeAPI,  constdef.Sell, "singal_test_53", "BTC")
		time.Sleep(time.Second*2)
	}
}

func TestZbLive_CalculateTradeStatus(t *testing.T) {
	zbTradeAPI := NewTradeAPI()
	zbTradeAPI.crossTradePairs =  []string{"btcqc", "ethqc", "ltcqc"}
	zbTradeAPI.RealTimeData = data.InitZbLiveData()
	CalculateTradeStatus(zbTradeAPI, constdef.Buy, "BTC")
}
