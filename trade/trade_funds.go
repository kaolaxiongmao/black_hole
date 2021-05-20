package trade

import (
	"black_hole/constdef"
	"black_hole/dal"
	"black_hole/utils"
	"context"
	"fmt"
	"log"
)

//仓位 需要借qc
func NeedLoadInQC(zbTradeAPI *ZbLive, coinName string, levers float64) float64 {
	crossAsset := zbTradeAPI.accountInfo.CrossAsset[coinName]
	virtualAvalivableNum := crossAsset.VirtualTotalNum
	virtualLoadIn := crossAsset.VirtualLoadIn
	nowPrice := zbTradeAPI.RealTimeData.GetDepthData().Buys[0][0]
	moneyAvalivableNum := crossAsset.MoneyTotalNum
	moneyLoadIn := crossAsset.MoneyLoadIn
	netAsset := (virtualAvalivableNum-virtualLoadIn)*nowPrice + (moneyAvalivableNum - moneyLoadIn)
	log.Println("[NeedLoadInQC].")
	loadInAsset := virtualLoadIn*nowPrice + moneyLoadIn
	NeedLoadInQC := netAsset*levers - loadInAsset
	return NeedLoadInQC
}

// 计算收益
func CalculateTradeStatus(zbTradeAPI *ZbLive, tradeType constdef.TradeType, coinName string) error {
	latestSignalKey, err := dal.GetLatestSinglKey(coinName)
	if err != nil {
		log.Fatalf("[CloseEmptyOrder.GetLatestSinglKey]. err: %v", err)
		return err
	}
	tradePairInfo, err := dal.GetTradePairOrderInfo(latestSignalKey)
	if err != nil {
		log.Fatalf("[CloseEmptyOrder.GetTradePairOrderInfo]. err. err: %v", err)
		return err
	}

	buyTotal := 0.0
	buyAmount := 0.0
	sellTotal := 0.0
	sellAmount := 0.0
	income := 0.0
	for _, order := range tradePairInfo.OrderMap {
		if order.SellOrBuy == int(constdef.Buy) {
			buyTotal = buyTotal + order.DealAmount*order.OrderPrice
			buyAmount = buyAmount + order.DealAmount
		} else {
			sellTotal = sellTotal + order.DealAmount*order.OrderPrice
			sellAmount = sellAmount + order.DealAmount
		}
	}
	log.Printf("[CalculateTradeStatus]: %v", tradePairInfo.OrderMap)
	info, err := dal.GetLoanInfo(latestSignalKey)
	if err != nil {
		return err
	}
	var incomePercent float64
	str := ""
	if tradeType == constdef.Sell {
		income = sellTotal - buyTotal - (sellTotal+buyTotal)*constdef.TradeFeed - info.LoanNum*zbTradeAPI.RealTimeData.GetDepthData().Buys[0][0]
		incomePercent = income / sellTotal*100
		str = str + fmt.Sprintf("✅平空相关信息如下: \n")
		// 平空计算收益
		str = str + fmt.Sprintf("✅借贷币种:【%s】, 借币的利息:%fQC\n", info.LoanCoin, info.LoanNum)
		str = str + fmt.Sprintf("✅卖出总QC:%f, 卖出平均价格:%f\n", sellTotal, sellTotal/sellAmount)
		str = str + fmt.Sprintf("✅买入个数:%f, 买入总qc:%f, 买入平均价格:%f\n", buyAmount, buyTotal, buyTotal/buyAmount)
		str = str + fmt.Sprintf("✅本次收益:%fQC, 本次收益率:%f,本次手续费%f, 本次借贷利息%f%s\n", income, incomePercent, (sellTotal+buyTotal)*constdef.TradeFeed, info.LoanNum*zbTradeAPI.RealTimeData.GetDepthData().Buys[0][0], "QC")
	} else {
		income = sellTotal - buyTotal - (sellTotal+buyTotal)*constdef.TradeFeed - info.LoanNum
		incomePercent = income / buyTotal*100

		// 平多计算收益
		str = str + fmt.Sprintf("✅平多相关信息如下: \n")
		str = str + fmt.Sprintf("✅借贷币种:【%s】借入的利息:%fQC\n", info.LoanCoin, info.LoanNum)
		str = str + fmt.Sprintf("✅买入总QC:%f, 买入平均价格:%f\n", buyTotal, buyTotal/buyAmount)
		str = str + fmt.Sprintf("✅卖出个数:%f, 卖出总qc:%f, 卖出平均价格:%f\n", sellAmount, sellTotal, sellTotal/sellAmount)
		str = str + fmt.Sprintf("✅本次收益:%fQC, 本次收益率:%f,本次手续费%f, 本次借贷利息%f%s\n", income, incomePercent, (sellAmount+buyAmount)*constdef.TradeFeed, info.LoanNum, "QC")
	}
	//收益率放入redis中
	err = dal.SetIncomePercent(latestSignalKey, incomePercent)
	if err != nil {
		log.Fatalf("[SetIncomePercent] err %v", err)
		return err
	}
	utils.SendLarkNoticeToDefaultUser(context.Background(), str)
	return nil
}
