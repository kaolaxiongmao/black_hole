package trade

import (
	"black_hole/constdef"
	"black_hole/dal"
	"black_hole/model"
	"black_hole/utils"
	"context"
	"fmt"
	"log"
	"time"
)

//开多单
//开空单
func OpenEmptyOrder(zbTradeAPI *ZbLive, singalKey string, coinName string) {
	ctx := context.Background()
	status, err := dal.GetLatestSingalStatus(coinName)
	if err != nil {
		utils.SendLarkNoticeToDefaultUser(ctx, "❌设置状态机状态失败")
		log.Fatalf("[OpenEmptyOrder.GetLatestSingalStatus] redis error")
		return
	}
	switch status {
	case constdef.TradeOrderFinish:
		//记得删除状态，赋予新状态
		err = dal.DelLatestSingalStatus(coinName)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌删除状态机状态失败")
		}
		utils.SendLarkNoticeToDefaultUser(ctx, "✅数据准备好 开始开空")
	case constdef.OpenOrderUninit:
		log.Print("[OpenEmptyOrder] OpenOrderUninit status")
		if status == constdef.OpenOrderUninit {
			if err := dal.SetLatestSingalKey(coinName, singalKey); err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, "❌设置交易key失败")
				return
			}
			if err := dal.SetLatestSingalStatus(coinName, int(constdef.OpenLoadIn)); err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, "❌设置交易状态失败")
				return
			}
			utils.SendLarkNoticeToDefaultUser(ctx, "✅开空：准备开始借款")
		}
	case constdef.OpenLoadIn:
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅开空：准备借入%v", coinName))
		log.Printf("[OpenEmptyOrder] OpenLoadIn status")
		if err := zbTradeAPI.GetCrossAsset(); err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("开空：❌获取杠杆账户失败 err%v", err))
			log.Fatalf("[OpenEmptyOrder] err.")
			return
		}
		//查看需要借多少个币
		virtualLoadIn := zbTradeAPI.accountInfo.CrossAsset[coinName].VirtualLoadIn
		qcTotalNum := zbTradeAPI.accountInfo.CrossAsset[coinName].NetQC
		nowPrice := zbTradeAPI.RealTimeData.GetDepthData().Buys[0][0]
		readyLoadNum := (qcTotalNum/nowPrice)*zbTradeAPI.emptyLever - virtualLoadIn
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅开空: 需要借币%f%s", readyLoadNum, coinName))
		log.Print("[OpenEmptyOrder] qcTotalNum: %v, needLoadIn: %v", qcTotalNum, readyLoadNum)
		if readyLoadNum > 0 {
			//算出需要借多少钱
			if err := zbTradeAPI.DoCrossLoan(readyLoadNum, coinName, constdef.CrossMarketMap[coinName]); err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌开空：借币失败 err:%v", err))
				log.Fatalf("[OpenEmptyOrder.DoCrossLoan]. err: %v", err.Error())
				return
			} else {
				err = dal.SetLatestSingalStatus(coinName, int(constdef.OpenTradeOrder))
				if err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, "❌开空：设置状态失败")
					return
				}
			}
		}
		err = dal.SetLatestSingalStatus(coinName, int(constdef.OpenTradeOrder))
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌开空：设置状态失败")
			return
		}

	case constdef.OpenTradeOrder:
		log.Print("[OpenEmptyOrder] OpenTradeOrder status.")
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅开空：卖出借入的%s", coinName))
		//挂单
		if err := zbTradeAPI.GetCrossAsset(); err != nil {
			return
		}
		tradePairInfo, err := dal.GetTradePairOrderInfo(singalKey)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌开空：获取杠杆账户失败 err %v", err))
			log.Fatalf("[OpenEmptyOrder.GetTradePairOrderInfo]. err. err: %v", err)
			return
		}
		if tradePairInfo.NowTradeId == "" {
			virtualTotalNum := zbTradeAPI.accountInfo.CrossAsset[coinName].VirtualTotalNum
			if virtualTotalNum < constdef.MinCoinUnit[coinName] {
				err = dal.SetLatestOpertionKey(coinName, int(constdef.Sell))
				if err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, "❌开空：设置最近操作失败")
					return
				}
				err = dal.SetLatestSingalStatus(coinName, int(constdef.OpenTradeOrderFinish))
				if err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, "❌开空：设置状态机状态失败")
					return
				}
				utils.SendLarkNoticeToDefaultUser(ctx, "✅空单： 空单成交完成 但是数量不够，需要check!!")
				return
			}
			//TODO:如果redis没有存储，那也要去最近挂单历史找出有没有挂单的内容, 有可能中间直接停掉了
			//这里是全仓交易
			log.Printf("[OpenEmptyOrder]. virtualTotalNum: %v", virtualTotalNum)
			id, orderPrice, orderNum, err := zbTradeAPI.SetOrder(int(constdef.Sell), coinName, virtualTotalNum, int(constdef.Futures))
			if err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌开空：下单失败 err %v", err))
				log.Fatalf("[OpenEmptyOrder.SetOrder]. err: %v", err.Error())
				return
			}

			if err := dal.SetNewOrder(singalKey, id, orderPrice, orderNum, int(constdef.Sell)); err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, "❌开空：设置卖出订单信息失败")
				log.Fatalf("[OpenEmptyOrder.SetNewOrder]. err: %v", err.Error())
				return
			}
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅开空: 订单卖出成功: 订单号%v, 卖出个数%f%s, 卖出价格%f", id, orderNum, coinName, orderPrice))

			//提前返回 等待一个tick检查订单状态
			return
		}

		//检查订单状态，如果超过阈值，直接就取消订单
		//查询状态
		itemStatus, err := zbTradeAPI.QueryOrderStatus(constdef.CrossMarketMap[coinName], tradePairInfo.NowTradeId)
		if err != nil {
			log.Fatalf("[OpenEmptyOrder.QueryOrderStatus]. err: %v", err)
			return
		}
		orderItem := model.OrderItem{
			TradeStatus:         int(itemStatus.Status),
			OrderPrice:          itemStatus.Price,
			OrderAmount:         itemStatus.OrderAmount,
			DealAmount:          itemStatus.DealAmount,
			StartOrderTimeStamp: itemStatus.Date,
			EndOrderTimeStamp:   time.Now().Unix(),
			SellOrBuy:           itemStatus.Type,
		}
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅订单号%v: 下单数量%f, 成交数量%f, 下单价格%f, 是否成功%v", tradePairInfo.NowTradeId, orderItem.OrderAmount, orderItem.DealAmount, orderItem.OrderPrice, itemStatus.Status == constdef.Success))

		err = dal.SetOrderInfo(singalKey, tradePairInfo.NowTradeId, orderItem)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌开空：设置订单信息失败")
			return
		}
		if itemStatus.Status == constdef.Success {
			err = dal.RemoveOrderInfo(singalKey)
			if err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, "❌开空：清空订单失败")
				return
			}
			err = dal.SetLatestOpertionKey(coinName, int(constdef.Sell))
			if err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, "❌开空： 设置最近操作失败")
				return
			}
			err = dal.SetLatestSingalStatus(coinName, int(constdef.OpenTradeOrderFinish))
			if err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, "❌开空：设置状态机状态失败")
				return
			}
			utils.SendLarkNoticeToDefaultUser(ctx, "✅开空单成功，所有卖单已成交")
			utils.SendLarkNoticeToDefaultUser(ctx, "✅空单成交完成")
			BroadAsset(zbTradeAPI)

		} else {
			if time.Now().Unix()-tradePairInfo.OrderMap[tradePairInfo.NowTradeId].StartOrderTimeStamp > constdef.SetOrderWaitSecond {
				utils.SendLarkNoticeToDefaultUser(ctx, "❌开空：空单挂单超时未完成，重新设置")
				//取消订单
				if err := zbTradeAPI.CancelOrder(constdef.CrossMarketMap[coinName], tradePairInfo.NowTradeId); err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌开空：取消订单失败，订单号:%v", tradePairInfo.NowTradeId))
					log.Fatalf("[CancelOrder]. cancelOrder error. err: %v", err)
					return
				}
				err = dal.RemoveOrderInfo(singalKey)
				if err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, "❌开空：清除订单信息失败")
				}
				utils.SendLarkNoticeToDefaultUser(ctx, "✅开空：开始重新挂单")
				return
			}
		}
	case constdef.OpenTradeOrderFinish:
		log.Fatalf("[OpenEmptyOrder.OpenTradeOrderFinish] Finish.")
		return
	default:
		return
	}
}

//平空单
func CloseEmptyOrder(zbTradeAPI *ZbLive, coinName string, newSingalKey string) {
	ctx := context.Background()
	status, err := dal.GetLatestSingalStatus(coinName)
	if err != nil {
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平空：获取redis失败 %v", err))
		log.Fatalf("[CloseEmptyOrder] redis error")
		return
	}
	log.Print("[CloseEmptyOrder]: status:%v", status)
	latestSingalKey, err := dal.GetLatestSinglKey(coinName)
	if err != nil {
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平空：获取redis失败 %v", err))
		log.Fatalf("[CloseEmptyOrder]. latestSingalKey: %v", latestSingalKey)
		return
	}
	log.Print("[CloseEmptyOrder]: latestSingalKey:%v", latestSingalKey)
	if latestSingalKey == newSingalKey {
		return
	}
	switch status {
	case constdef.OpenTradeOrderFinish:
		log.Print("[CloseEmptyOrder] OpenTradeOrderFinish.")
		err = dal.SetLatestSingalStatus(coinName, int(constdef.CloseTradeOrder))
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平空：设置状态机status失败 err:%v", err))
			log.Fatalf("[dal.SetLatestSingalStatus]. err: %v", err)
			return
		}
		utils.SendLarkNoticeToDefaultUser(ctx, "✅开始平空")

	case constdef.CloseTradeOrder:
		log.Print("[CloseEmptyOrder] CloseTradeOrder.")
		if err := zbTradeAPI.GetCrossAsset(); err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌平空：获取杠杆账号失败")
			log.Fatalf("[CloseEmptyOrder.GetCrossAsset]. err: %v", err.Error())
			return
		}
		crossAsset := zbTradeAPI.accountInfo.CrossAsset[coinName]
		//还币这里要给一个最小单位
		needBuyCoinNum := (crossAsset.VirtualFeed + constdef.MinCoinUnit[coinName] + crossAsset.VirtualLoadIn - crossAsset.VirtualTotalNum) * (1 + constdef.TradeFeed)
		//计算最小单
		needQcNum := zbTradeAPI.RealTimeData.GetDepthData().Sells[0][0] * needBuyCoinNum
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅平空：需要买币:%f%s,需要qc:%f", needBuyCoinNum, coinName, needQcNum))
		log.Print("[CloseEmptyOrder]. needBuyCoinNum: %v, needQcNum: %v", needBuyCoinNum, needQcNum)
		var emptyId string
		latestSignalKey, err := dal.GetLatestSinglKey(coinName)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平空：获取最近一次交易key失败 err:%v", err))
			log.Fatalf("[CloseEmptyOrder.GetLatestSinglKey]. err: %v", err)
			return
		}
		tradePairInfo, err := dal.GetTradePairOrderInfo(latestSignalKey)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平空：获取交易对信息失败 err:%v", err))
			log.Fatalf("[CloseEmptyOrder.GetTradePairOrderInfo]. err. err: %v", err)
			return
		}
		if tradePairInfo.NowTradeId == "" {
			utils.SendLarkNoticeToDefaultUser(ctx, "✅平空： 开始委托挂单")
			if needBuyCoinNum > constdef.MinCoinUnit[coinName] {
				//挂单
				var orderErr error
				var orderPrice, orderNum float64
				emptyId, orderPrice, orderNum, orderErr = zbTradeAPI.SetOrder(int(constdef.Buy), coinName, needQcNum, int(constdef.Futures))
				if orderErr != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平空：买入失败 err:%v", orderErr))
					log.Fatalf("[CloseEmptyOrder.SetOrder]. err: %v", orderErr.Error())
					return
				}
				if err := dal.SetNewOrder(latestSignalKey, emptyId, orderPrice, orderNum, int(constdef.Buy)); err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平空：设置订单信息失败 err:%v", err))
					log.Fatalf("[CloseEmptyOrder.SetNewOrder]. err: %v", err.Error())
					return
				}
				utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅平空：买入平空挂单成功，订单号:%s,开单价格:%f,挂单数量:%f", emptyId, orderPrice, orderNum))

				//提前返回 等待一个tick检查订单状态
				return
			} else {
				err = dal.SetLatestSingalStatus(coinName, int(constdef.ClosePayBack))
				if err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平空：设置状态机status失败 err:%v", err))
					log.Fatalf("[dal.SetLatestSingalStatus]. err: %v", err)
					return
				}
				utils.SendLarkNoticeToDefaultUser(ctx, "✅平空:买入需要还的币成功")
				return
			}
		}
		//检查订单状态，如果超过阈值，直接就取消订单
		//查询状态
		itemStatus, err := zbTradeAPI.QueryOrderStatus(constdef.CrossMarketMap[coinName], tradePairInfo.NowTradeId)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平空：获取订单信息失败 err:%v", err))
			log.Fatalf("[OpenOrder.QueryOrderStatus]. err: %v", err)
			return
		}
		orderItem := model.OrderItem{
			TradeStatus:         int(itemStatus.Status),
			OrderPrice:          itemStatus.Price,
			OrderAmount:         itemStatus.OrderAmount,
			DealAmount:          itemStatus.DealAmount,
			StartOrderTimeStamp: itemStatus.Date,
			EndOrderTimeStamp:   time.Now().Unix(),
			SellOrBuy:           itemStatus.Type,
		}
		err = dal.SetOrderInfo(latestSignalKey, tradePairInfo.NowTradeId, orderItem)
		if err != nil {
			log.Fatalf("[dal.SetOrderInfo]. err: %v", err)
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平空：更新订单信息失败 err:%v", err))
			return
		}
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅平空：获取买入订单状态，开单价格:%f,下单个数:%f,成交个数:%f，是否完全成交%v", orderItem.OrderPrice, orderItem.OrderAmount, orderItem.DealAmount, itemStatus.Status == constdef.Success))

		if itemStatus.Status == constdef.Success {
			err = dal.RemoveOrderInfo(latestSignalKey)
			if err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平空：删除当前订单id失败 err:%v", err))
				log.Fatalf("[dal.RemoveOrderInfo]. err: %v", err)
				return
			}
			err = dal.SetLatestSingalStatus(coinName, int(constdef.ClosePayBack))
			if err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平空：设置状态机status失败 err:%v", err))
				log.Fatalf("[dal.SetLatestSingalStatus]. err: %v", err)
				return
			}
			utils.SendLarkNoticeToDefaultUser(ctx, "✅买入平空交易成功")

		} else {
			if time.Now().Unix()-tradePairInfo.OrderMap[tradePairInfo.NowTradeId].StartOrderTimeStamp > constdef.BuyCoinMaxWaitSecond {
				utils.SendLarkNoticeToDefaultUser(ctx, "❌平空：买入平空挂单超时未完成，重新设置")
				//取消订单
				if err != zbTradeAPI.CancelOrder(constdef.CrossMarketMap[coinName], tradePairInfo.NowTradeId) {
					utils.SendLarkNoticeToDefaultUser(ctx, "❌平空：取消订单失败")
					log.Fatalf("[CancelOrder]. cancelOrder error. err: %v", err)
					return
				}
				err = dal.RemoveOrderInfo(latestSingalKey)
				if err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, "❌平空：清除订单信息失败")
				}
				utils.SendLarkNoticeToDefaultUser(ctx, "✅平空：超时重新设置买入平空挂单")
				return
			}
		}
	case constdef.ClosePayBack:
		// 记录一下还款币种，还款个数
		latestSignalKey, err := dal.GetLatestSinglKey(coinName)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌平空：获取上次交易key失败")
			log.Fatalf("[CloseEmptyOrder.GetLatestSinglKey]. err: %v", err)
			return
		}
		err = dal.SetLoanInfo(latestSignalKey, coinName, zbTradeAPI.accountInfo.CrossAsset[coinName].VirtualFeed)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平空：设置借款利息失败 err:%v", err))
			log.Fatalf("[dal.SetLoanInfo]. err: %v", err)
			return
		}
		// 还款
		err = zbTradeAPI.DoCrossRepay(coinName, constdef.CrossMarketMap[coinName])
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平空：还%f%s失败 err:%v", zbTradeAPI.accountInfo.CrossAsset[coinName].VirtualLoadIn, coinName, err))
			log.Fatalf("[CloseEmptyOrder.DoCrossRepay]. err: %v", err)
			return
		}
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅平空：还%f%s成功", zbTradeAPI.accountInfo.CrossAsset[coinName].VirtualLoadIn, coinName))

		err = dal.SetLatestSingalStatus(coinName, int(constdef.CloseTradeOrderFinish))
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌设置状态机status失败 err:%v", err))
			log.Fatalf("[CloseEmptyOrder.SetLatestSingalStatus]. err: %v", err)
			return
		}
		utils.SendLarkNoticeToDefaultUser(ctx, "✅平空：平空还款完成")

	case constdef.CloseTradeOrderFinish:
		log.Print("[CloseEmptyOrder.CloseTradeOrderFinish]")
		err = dal.SetLatestSingalStatus(coinName, int(constdef.TradeOrderEndHandler))
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌设置状态机status失败 err:%v", err))
			log.Fatalf("[CloseEmptyOrder.SetLatestSingalStatus]. err: %v", err)
			return
		}
		utils.SendLarkNoticeToDefaultUser(ctx, "✅平空： 准备计算利润")
	case constdef.TradeOrderEndHandler:
		log.Print("[CloseEmptyOrder.TradeOrderEndHandler]")
		err := CalculateTradeStatus(zbTradeAPI, constdef.Sell, coinName)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌计算利润失败 err:%v", err))
			log.Fatalf("[CalculateTradeStatus] err %v", err)
			return
		}
		err = dal.DelLatestOperation(coinName)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌清理最近一个操作错误。err:%v", err))
			log.Fatalf("[CloseEmptyOrder.DelLatestOperation]. err: %v", err)
			return
		}
		err = dal.SetLatestSingalStatus(coinName, int(constdef.TradeOrderFinish))
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌设置状态机status失败 err:%v", err))
			log.Fatalf("[CloseEmptyOrder.SetLatestSingalStatus]. err: %v", err)
			return
		}
		utils.SendLarkNoticeToDefaultUser(ctx, "✅平空：平空结束，计算利润完成， 订单完成")
		BroadAsset(zbTradeAPI)
	case constdef.TradeOrderFinish:
		log.Print("[CloseEmptyOrder.TradeOrderFinish]")
		return
	default:
		return
	}
}

