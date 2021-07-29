package backtest

import (
	"black_hole/chart"
	"black_hole/constdef"
	"black_hole/model"
	"black_hole/strategy"
	"github.com/markcheno/go-talib"
	"log"
	"time"
)

var empty = map[float64]float64{
	0.5: -95,
	1:   -90,
	1.5: -56.6,
	2:   -40,
	3:   -23.3,
}
var mutipleMap = map[float64]float64{
	1:   -45,
	1.5: -34,
	2:   -26.6,
	3:   -17.5,
}

//回测模型主主体框架
//通过传入策略，时间范围，生成一个列表
func BackTest(tradeStrategy *strategy.Strategy, emptySingle float64, mutipleSingle float64, orderFeed float64, loadInFeed float64) {
	time.Sleep(time.Second * 3)
	structDatas := tradeStrategy.Params.KlineDatas
	startIndex := 0
	var buyPrice float64
	var sellPrice float64
	var startMoney float64 = 1
	var buyDate string
	var sellDate string
	var nowProfit float64
	//正确做多次数
	correctBuyCnt := 0
	correctBuyCntOneBar := 0
	//做多次数
	buyCnt := 0
	//正确做空次数
	correctSellCnt := 0
	//做空次数
	sellCnt := 0
	correctSellCntOneBar := 0
	startStrategy := constdef.StandBy
	// 手续费
	tradeFee := 0.0
	// 借贷费用
	loanFee := 0.0
	var holdDay int64 = 0 //持仓日期
	var backTestDatas []model.BackTestDataElement
	for i := 29; i <= len(structDatas)-1; i++ {
		tradeStrategy.Params.KlineDatas = structDatas[startIndex : i+1]
		startIndex++
		buyOrSell := tradeStrategy.Tick(startStrategy)
		switch buyOrSell {
		case constdef.StandBy:
			holdDay++ //加入持仓日期
			if startStrategy == constdef.Buy {
				checkIsLiquidate(structDatas[i].TimeStampStr, structDatas[i].LowPrice, buyPrice, mutipleSingle, holdDay, loadInFeed, false)
				backTestDatas = append(backTestDatas, genBackTestGraphElement(buyPrice, structDatas[i], false, mutipleSingle, startMoney))
				if holdDay == 3 {
					if buyPrice - structDatas[i].ClosePrice < 0 {
						correctBuyCntOneBar++
					}
				}
			} else if startStrategy == constdef.Sell {
				checkIsLiquidate(structDatas[i].TimeStampStr, structDatas[i].HighPrice, sellPrice, emptySingle, holdDay, loadInFeed, true)
				backTestDatas = append(backTestDatas, genBackTestGraphElement(sellPrice, structDatas[i], true, mutipleSingle, startMoney))
				if holdDay == 3 {
					if sellPrice - structDatas[i].ClosePrice > 0 {
						correctSellCntOneBar++
					}
				}
			}
		case constdef.Buy:
			startStrategy = constdef.Buy
			nowProfit = 0
			buyDate = structDatas[i].TimeStampStr
			buyPrice = structDatas[i].ClosePrice
			if sellPrice > 0 {
				//平空 然后买入
				backTestDatas = append(backTestDatas, genBackTestGraphElement(sellPrice, structDatas[i], true, emptySingle, startMoney))
				loanFee += startMoney*float64(holdDay)*loadInFeed*emptySingle
				tradeFee += startMoney*orderFeed
				startMoney = startMoney*(sellPrice-buyPrice)/sellPrice*emptySingle + startMoney - startMoney*orderFeed - startMoney*float64(holdDay)*loadInFeed*emptySingle
				holdDay = 0
				nowProfit = (sellPrice - buyPrice) / sellPrice * emptySingle * 100
				log.Printf("[Order], 平空 BuyDate: %v, BuyPrice: %v", buyDate, buyPrice)
				log.Printf("[Order], 空单利润: %v, 当前总额: %v", nowProfit, startMoney)
				if nowProfit > 0 {
					correctSellCnt++
				}
			}
			if nowProfit > 30 {
				//空单利润高， 就不做多
				buyPrice = 0
				continue
			}
			log.Printf("[Order], 做多 BuyDate: %v, BuyPrice: %v", buyDate, buyPrice)
			//扣除手续费
			startMoney = startMoney - startMoney*orderFeed
			tradeFee+= startMoney*orderFeed
			buyCnt++
			holdDay++
		case constdef.Sell:
			startStrategy = constdef.Sell
			nowProfit = 0
			sellPrice = structDatas[i].ClosePrice
			sellDate = structDatas[i].TimeStampStr
			if buyPrice > 0 {
				backTestDatas = append(backTestDatas, genBackTestGraphElement(buyPrice, structDatas[i], false, mutipleSingle, startMoney))
				loanFee += startMoney*float64(holdDay)*loadInFeed*emptySingle
				tradeFee += startMoney*orderFeed
				startMoney = startMoney*(sellPrice-buyPrice)/buyPrice*(mutipleSingle+1) + startMoney - startMoney*float64(holdDay)*loadInFeed*mutipleSingle
				holdDay = 0
				//扣除手续费
				startMoney = startMoney - startMoney*orderFeed
				sellDate = structDatas[i].TimeStampStr
				nowProfit = (sellPrice - buyPrice) / buyPrice * (mutipleSingle + 1) * 100
				log.Printf("[Order], 平多 SellDate: %v, SellPrice: %v", sellDate, sellPrice)
				log.Printf("[Order], 多单利润: %v, 当前总额: %v", nowProfit, startMoney)
				if nowProfit > 0 {
					correctBuyCnt++
				}
			}
			if nowProfit > 30 {
				//多单单笔利润高，不做空
				sellPrice = 0
				continue
			}
			log.Printf("[Order], 做空 SellDate: %v, SellPrice: %v", sellDate, sellPrice)
			//扣除手续费
			tradeFee += startMoney*orderFeed
			startMoney = startMoney - startMoney*orderFeed
			holdDay++
			sellCnt++
		}
	}
	log.Printf("做空胜率： %v, 做多胜率: %v, 做空一个bar胜率： %v, 做多一个bar胜率： %v", float64(correctSellCnt)/float64(sellCnt)*100, float64(correctBuyCnt)/float64(buyCnt)*100, float64(correctSellCntOneBar)/float64(sellCnt)*100, float64(correctBuyCntOneBar)/float64(buyCnt)*100)
	log.Printf("交易费率: %v, 借贷费率:%v", tradeFee, loanFee)
	chart.StartHttpServer("宙斯", "btc", backTestDatas)
	time.Sleep(time.Second)
}

//生成图标
func genBackTestGraphElement(openTradePrice float64, kLineBar model.KLineBar, isSell bool, mutiple float64, totalMoney float64) model.BackTestDataElement {

	if !isSell {
		if openTradePrice > 0 {
			totalMoney = totalMoney*(kLineBar.ClosePrice-openTradePrice)/openTradePrice*(mutiple+1) + totalMoney
		}
	} else {
		if openTradePrice > 0 {
			totalMoney = totalMoney*(openTradePrice-kLineBar.ClosePrice)/openTradePrice*mutiple + totalMoney
		}
	}
	element := model.BackTestDataElement{
		Date:        kLineBar.TimeStampStr,
		OpenPrice:   kLineBar.OpenPrice,
		ClosePrice:  kLineBar.ClosePrice,
		HightPrice:  kLineBar.HighPrice,
		LowPrice:    kLineBar.LowPrice,
		TotalProfit: totalMoney,
	}
	return element
}

func checkIsLiquidate(timeStampStr string, barPrice float64, orderPrice float64, mutiple float64, holdDay int64, loadInFeed float64, isSell bool) {
	if isSell {
		if barPrice > 0 && orderPrice > 0 {
			log.Printf("[checkIsLiquidate] 检查空单爆仓： barPrice: %v, orderPrice: %v, timeStampStr: %v, 盈利率: %v", barPrice, orderPrice, timeStampStr, (orderPrice-barPrice)/orderPrice*float64(mutiple)*100)
			if (orderPrice-barPrice)/orderPrice*float64(mutiple)*100-(float64(holdDay)*loadInFeed*100) < empty[mutiple] {
				log.Fatalf("[checkIsLiquidate]. 空单爆仓. timeStampStr: %v, maxPrice: %v, orderPirce: %v, 当前盈利：%v, 爆仓阈值： %v", timeStampStr, barPrice, orderPrice, (orderPrice-barPrice)/orderPrice*float64(mutiple)*100, empty[mutiple])
				time.Sleep(time.Second)
				panic("baocang")
			}
		}
	}
	if !isSell {
		if barPrice > 0 && orderPrice > 0 {
			log.Printf("[checkIsLiquidate] 检查多单爆仓： barPrice: %v, orderPrice: %v, timeStampStr: %v, 盈利率: %v", barPrice, orderPrice, timeStampStr, (barPrice-orderPrice)/orderPrice*float64(mutiple)*100)
			if (barPrice-orderPrice)/orderPrice*float64(mutiple)*100-(float64(holdDay)*loadInFeed*100) < mutipleMap[mutiple] {
				log.Fatalf("[checkIsLiquidate]. 多单爆仓. timeStampStr: %v, maxPrice: %v, orderPirce: %v", timeStampStr, barPrice, orderPrice)
				time.Sleep(time.Second)
				panic("baocang")
			}
		}
	}
}

func DrawBias(structDatas []model.KLineBar){
	startIndex := 0
	var elements [] model.BiasDataElement
	for i := 30; i <= len(structDatas)-1; i++{
		KlineDatas := structDatas[startIndex : i+1]
		var closePrices []float64
		for _, barData := range KlineDatas {
			closePrices = append(closePrices, barData.ClosePrice)
		}
		ma := talib.Sma(closePrices, 10)
		nowIndex := len(KlineDatas)-1
		var bia float64
		bia = (KlineDatas[nowIndex].ClosePrice - ma[nowIndex]) /ma[nowIndex]
		if bia > 0.2{
			log.Printf("乖离率>20:%v", KlineDatas[nowIndex].TimeStampStr)
		}
		startIndex++
		elements = append(elements, GenBiaElement(bia, KlineDatas[nowIndex].TimeStampStr))
	}
	chart.StartBiasHttpServer("10日乖离率", "btc", elements)
}

func GenBiaElement(bias float64, data string) model.BiasDataElement {
	return model.BiasDataElement{
		Data: data,
		Bias: bias,
	}
}