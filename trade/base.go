package trade

import "black_hole/constdef"

//存一个最近的单号
func OpenOrder(zbTradeAPI *ZbLive, singal constdef.TradeType, singalKey string, coinName string) {
	if singal == constdef.Buy {
		//平空
		CloseEmptyOrder(zbTradeAPI, coinName, singalKey)
		//开多
		OpenMutipleOrder(zbTradeAPI, singalKey, coinName)
	} else if singal == constdef.Sell {
		//平多
		CloseMutipleOrder(zbTradeAPI, coinName, singalKey)
		//开空
		OpenEmptyOrder(zbTradeAPI, singalKey, coinName)
	}
}

