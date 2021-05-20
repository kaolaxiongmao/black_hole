package constdef

import (
	"black_hole/conf"
	"time"
)

const (
	TradeZbUrl = "https://trade.zb.center/api"
	DataZbUrl  = "http://api.zb.center/data/v1/"
)

//market名字
var MarketMap = map[string]string{
	"BTC": "btc_qc",
	"LTC": "ltc_qc",
	"ETH": "eth_qc",
}

//杠杆market市场
var CrossMarketMap = map[string]string{
	"LTC": "ltcqc",
	"BTC": "btcqc",
}
var BarLine = []string{"1day", "15min"}

//交易的最小单位 qc
var MinCoinUnit = map[string]float64{
	"ETH": 0.001,
	"BTC": 0.0001,
	"XRP": 1,
	"EOS": 0.01,
	"LTC": 0.1,
}

//最少50个qc发起订单
var MinQcUnit = float64(50)

//买入
type TradeType int

const (
	Buy  = TradeType(1)
	Sell = TradeType(0)
	//这个状态的意思不买也不卖
	StandBy = TradeType(2)
)

type OrderStatus int

const (
	Success    = OrderStatus(2) //成功
	Processing = OrderStatus(3) //正在处理中
	Canceled   = OrderStatus(1) //取消
)
const (
	TradeFeed = 0.002 //交易费率
)

type AssetType int

const (
	//现货
	Goods = AssetType(0)
	//期货-逐仓
	Futures = AssetType(1)
)

//获取深度频率
const (
	DepthFreq = time.Second * 1
	KFreq     = time.Minute * 5
)


var FocusCoins = []string{"QC", "ETH", "BTC"}

var LeverCrossCoins = []string{"btcqc", "ethqc", "ltcqc"}

type TradeOrderStatus int

const (
	OpenOrderUninit       = TradeOrderStatus(0) //未开单
	OpenLoadIn            = TradeOrderStatus(1) //借款状态
	OpenTradeOrder        = TradeOrderStatus(2) //开单委托挂单状态
	OpenTradeOrderFinish  = TradeOrderStatus(3) //开单委托挂单成功
	CloseTradeOrder       = TradeOrderStatus(4) //平仓挂单
	CloseTradeOrderFinish = TradeOrderStatus(6) //平仓完成
	ClosePayBack          = TradeOrderStatus(5) //平仓还款
	TradeOrderEndHandler  = TradeOrderStatus(7) //处理结束 包括利润计算等
	TradeOrderFinish      = TradeOrderStatus(8) //全部结束 收尾
)

func GetCurMarket() string {
	config := conf.GetConfig()
	return MarketMap[config.CurCurrency]
}
const (
	BuyCoinMaxWaitSecond = 60 //买币最大时长
	SetOrderWaitSecond = 60 //挂单最长时间
)
