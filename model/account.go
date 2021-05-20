package model

import (
	"black_hole/constdef"
	"time"
)

//期货账户
type LeverStrAccountInfo struct {
	CAvailable     string `json:"cAvailable"` //数字货币当前含有的数量
	CLoanIn        string `json:"cLoanIn"`    //数字货币借入的数量
	COverdraft     string `json:"cOverdraft"` //数字货币利息
	FAvailable     string `json:"fAvailable"` //法币的总量
	FLoanIn        string `json:"fLoanIn"`    //法币的接入总量
	FOverDraft     string `json:"fOverdraft"` //法币利息
	CoinName       string `json:"cEnName"`
	TradePair      string `json:"key"`            //交易对
	UnwindPrice    string `json:"unwindPrice"`    // 平仓价格
	RepayLeverShow string `json:"repayLeverShow"` // 平仓风险
	NetQC          string `json:"netConvertCNY"`  //净资产
}

type LeverAccountInfo struct {
	VirtualTotalNum float64 //数字货币当前含有的数量
	VirtualLoadIn   float64 //数字货币借入的数量
	VirtualFeed     float64 //数字货币利息
	MoneyTotalNum   float64 //法币的总量 qc
	MoneyLoadIn     float64 //法币的借入总量
	MoneyFeed       float64 //法币利息
	CoinName        string
	UnwindPrice     float64 //平仓价格
	RepayLeverShow  string  //平仓风险
	NetQC           float64  //净资产
}

//账户相关
type AccountInfo struct {
	VirtualCurrency map[string]float64          //货币相关
	CrossAsset      map[string]LeverAccountInfo //杠杆账户
}
type Pairs []float64

//深度相关
type DepthData struct {
	TimeStamp      int64       `json:"timestamp"`
	Sells          [][]float64 `json:"asks"` //卖挂单
	Buys           [][]float64 `json:"bids"` //买账单
	LocalTimeStamp int64       //本地时钟
}

type KLineData struct {
	Data       [][]float64 `json:"data"`
	MoneyType  string      `json:"moneyType"`
	Symbol     string      `json:"symbol"`
	StructData []KLineBar
}

//k线所含的内容
type KLineBar struct {
	OpenPrice    float64 `json:"open_price"`
	ClosePrice   float64 `json:"close_price"`
	LowPrice     float64 `json:"low_price"`
	HighPrice    float64 `json:"high_price"`
	Volum        float64 `json:"volum"`
	TimeStamp    int64   `json:"time_stamp"`
	TimeStampStr string  `json:"time_stamp_str"`
	Date         time.Time
	CoinName     string
}

//交易
type OrderItem struct {
	TradeStatus         int     //买入状态
	OrderPrice          float64 //挂单买入价格
	OrderAmount         float64 //挂单买入数量
	DealAmount          float64 //成交的数量
	StartOrderTimeStamp int64   //开单时间戳
	EndOrderTimeStamp   int64   //关单时间戳
	SellOrBuy           int     //开空or开多
}

// 借款情况
type LoanInfo struct {
	LoanCoin string  // 借款币种
	LoanNum  float64 // 借款个数
}

//交易对
type TradePair struct {
	NowTradeId    string
	OrderMap      map[string]OrderItem //单map order_id -> info
	ProfitPercent float64              //利润百分比
}

type OrderItemStatus struct {
	Status      constdef.OrderStatus
	Price       float64
	OrderAmount float64
	DealAmount  float64
	Type        int
	Date        int64
}

