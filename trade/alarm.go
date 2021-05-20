package trade

import (
	"black_hole/conf"
	"black_hole/model"
	"black_hole/utils"
	"context"
	"fmt"
	"github.com/markcheno/go-talib"
	"log"
)

var coinMap = map[string]string{"BTC": "比特币", "LTC": "莱特币"}

func BroadAsset(zbTradeAPI *ZbLive) {
	if err := zbTradeAPI.GetCrossAsset(); err != nil {
		log.Fatalf("[OpenOrder.GetCrossAsset]. err: %v", err.Error())
		return
	}
	asset := zbTradeAPI.accountInfo.CrossAsset // 先取当前价格
	curBuyPrice := zbTradeAPI.RealTimeData.GetDepthData().Buys[0][0]
	curSellPrice := zbTradeAPI.RealTimeData.GetDepthData().Sells[0][0]
	c := conf.GetConfig().CurCurrency
	coinInfo := asset[c]
	qcLoan := coinInfo.MoneyFeed
	qcBorrow := coinInfo.MoneyLoadIn
	qcTotal := coinInfo.MoneyTotalNum
	virtualLoan := coinInfo.VirtualFeed
	virtualBorrow := coinInfo.VirtualLoadIn
	virtualTotal := coinInfo.VirtualTotalNum
	repayLevelShow := coinInfo.RepayLeverShow
	unwindPrice := coinInfo.UnwindPrice
	kLine := zbTradeAPI.RealTimeData.GetKData()
	structDatas := kLine["1day"].StructData
	ma5 := CalCurMa(5,  structDatas)
	ma20 := CalCurMa(20,structDatas)
	diffRate := (curBuyPrice - CalCurMa(10, structDatas))/CalCurMa(10,  structDatas)
	// 计算当前总借款
	str := ""
	str = str + fmt.Sprintf("✅当前持有币种:【%s】, 当前买一价格:%f, 当前卖一价格:%f\n", coinMap[c], curBuyPrice, curSellPrice)
	str = str + fmt.Sprintf("✅持有虚拟币个数:%f, 借虚拟币个数:%f, 虚拟币利息:%f\n", virtualTotal, virtualBorrow, virtualLoan)
	str = str + fmt.Sprintf("✅持有qc个数:%f, 借qc:%f, qc利息:%f\n", qcTotal, qcBorrow, qcLoan)
	// 计算风险率
	str = str + fmt.Sprintf("✅风险率:%s, 爆仓价格:%f\n", repayLevelShow, unwindPrice)
	str = str + fmt.Sprintf("✅仓位净值:%f\n", zbTradeAPI.accountInfo.CrossAsset[conf.GetConfig().CurCurrency].NetQC)
	str = str + fmt.Sprintf("✅当前5日平均线:%f, 20日平均线:%f\n", ma5, ma20)
	str = str + fmt.Sprintf("✅当前乖离率:%f\n", diffRate*100)
	utils.SendLarkNoticeToDefaultUser(context.Background(), str)
}

func BroadInit(zbTradeAPI *ZbLive) {
	str := ""
	str = str + fmt.Sprintf("✅初始化配置\n")
	str = str + fmt.Sprintf("✅当前【币种】%v\n", zbTradeAPI.currencyName)
	str = str + fmt.Sprintf("✅当前杠杆率【买入杠杆率】%v, 【卖出杠杆率】%v\n", zbTradeAPI.mutipleLever, zbTradeAPI.emptyLever)
	str = str + fmt.Sprintf("✅当前账号信息【accessKey】%v\n 【sercetKey】%v\n 【password】%v\n", zbTradeAPI.accessKey, zbTradeAPI.sercetKey, zbTradeAPI.passWord)
	str = str + fmt.Sprintf("✅【时间间隔]%v\n", zbTradeAPI.barTime)
	utils.SendLarkNoticeToDefaultUser(context.Background(), str)
}

func CalCurMa(ma int, structDatas []model.KLineBar)float64{
	var closePrices []float64
	for _, barData := range structDatas {
		closePrices = append(closePrices, barData.ClosePrice)
	}
	maPrice := talib.Sma(closePrices, ma)
	return maPrice[len(closePrices)-1]
}
