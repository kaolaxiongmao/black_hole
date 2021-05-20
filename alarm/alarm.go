package alarm

import (
	"black_hole/trade"
	"log"
)

func BroadAsset(zbTradeAPI *trade.ZbLive){
	if err := zbTradeAPI.GetCrossAsset(); err != nil {
		log.Printf("[OpenOrder.GetCrossAsset]. err: %v\n", err.Error())
		return
	}
}
