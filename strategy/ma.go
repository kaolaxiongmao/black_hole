package strategy

import (
	"black_hole/constdef"
	"black_hole/model"
	"github.com/markcheno/go-talib"
	"log"
)

//买入策略 一定是要有买入策略，才会有卖出策略
//双均线策略
//kLineDatas k线数据
//threshHold ma策略阈值
//slowBarNum  ma多少个Bar
//fastBarNum  ma多少个bar
func DoubleMaBuy(params *model.StrategyParms) constdef.TradeType {
	var closePrices []float64
	for _, barData := range params.KlineDatas {
		closePrices = append(closePrices, barData.ClosePrice)
	}
	maSlowPrice := talib.Sma(closePrices, params.SlowMaBars)
	maFastPrice := talib.Sma(closePrices, params.FastMaBars)
	nowMaIndex := len(closePrices) - 1
	if maFastPrice[nowMaIndex] > maSlowPrice[nowMaIndex] && maFastPrice[nowMaIndex-1] < maSlowPrice[nowMaIndex-1] {
		return constdef.Buy
	}
	if maFastPrice[nowMaIndex] > maSlowPrice[nowMaIndex] && closePrices[nowMaIndex] > maSlowPrice[nowMaIndex] && closePrices[nowMaIndex-1] < maSlowPrice[nowMaIndex-1] {
		//log.Print("[DoubleMaBuy]  Buy Signal, Date: %v, Reson: fast cross slow", params.KlineDatas[nowMaIndex].TimeStampStr)
		return constdef.Buy
	}
	return constdef.StandBy
}

//
func DoubleMaSell(params *model.StrategyParms) constdef.TradeType {
	var closePrices []float64
	for _, barData := range params.KlineDatas {
		closePrices = append(closePrices, barData.ClosePrice)
	}
	maSlowPrice := talib.Sma(closePrices, params.SlowMaBars)
	maFastPrice := talib.Sma(closePrices, params.FastMaBars)
	nowMaIndex := len(closePrices) - 1
	if maFastPrice[nowMaIndex] < maSlowPrice[nowMaIndex] && maFastPrice[nowMaIndex-1] > maSlowPrice[nowMaIndex-1]{
		log.Printf("[DoubleMaSell] Sell Signal, Date : %v, Reson: slow cross fast", params.KlineDatas[nowMaIndex].TimeStampStr)
		return constdef.Sell
	}
	//收盘跌倒了慢线以下 卖出
	if closePrices[nowMaIndex] < maSlowPrice[nowMaIndex] && closePrices[nowMaIndex-1] > maSlowPrice[nowMaIndex-1]{
		log.Printf("[DoubleMaSell] Sell Signal, Date : %v, Reson: close price low fast price", params.KlineDatas[nowMaIndex].TimeStampStr)
		return constdef.Sell
	}
	return constdef.StandBy
}

