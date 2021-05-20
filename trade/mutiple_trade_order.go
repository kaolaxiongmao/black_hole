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
func OpenMutipleOrder(zbTradeAPI *ZbLive, singalKey string, coinName string) {
	ctx := context.Background()
	status, err := dal.GetLatestSingalStatus(coinName)
	if err != nil {
		utils.SendLarkNoticeToDefaultUser(ctx, "❌开多：获取状态机状态失败")
		log.Fatalf("[OpenMutipleOrder.GetLatestSingalStatus] redis error")
		return
	}
	switch status {
	case constdef.TradeOrderFinish:
		//记得删除状态，赋予新状态
		err = dal.DelLatestSingalStatus(coinName)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌开多：删除状态机状态失败")
			log.Fatalf("[OpenMutipleOrder] DelLatestSingalStatus redis error")
			return
		}
		utils.SendLarkNoticeToDefaultUser(ctx, "✅开多：数据ready, 准备开多")
	case constdef.OpenOrderUninit:
		log.Print("[OpenMutipleOrder] OpenOrderUninit status")
		if status == constdef.OpenOrderUninit {
			if err := dal.SetLatestSingalKey(coinName, singalKey); err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌开多：设置交易key失败 err:%v", err))
				return
			}
			if err := dal.SetLatestSingalStatus(coinName, int(constdef.OpenLoadIn)); err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌开多：设置状态机状态失败 err:%v", err))
				return
			}
			utils.SendLarkNoticeToDefaultUser(ctx, "✅开多：进入开多借款状态")
		}
	case constdef.OpenLoadIn:
		log.Print("[OpenMutipleOrder] OpenLoadIn status")
		if err := zbTradeAPI.GetCrossAsset(); err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌开多：获取杠杆账号失败")
			log.Fatalf("[OpenMutipleOrder] err.")
			return
		}
		//每次都查询 查看是否已经借成功
		needLoadIn := NeedLoadInQC(zbTradeAPI, coinName, zbTradeAPI.mutipleLever)
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅开多：还需要借款:%fQC", needLoadIn))
		log.Printf("[OpenMutipleOrder] needLoadIn: %v\n", needLoadIn)
		if needLoadIn > 0 {
			//算出需要借多少钱
			if err := zbTradeAPI.DoCrossLoan(needLoadIn, "QC", constdef.CrossMarketMap[coinName]); err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌开多借款失败 err:%v", err))
				log.Fatalf("[OpenOrder.DoCrossLoan]. err: %v", err.Error())
				return
			} else {
				err = dal.SetLatestSingalStatus(coinName, int(constdef.OpenTradeOrder))
				if err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌开多： 设置状态机状态失败 err:%v", err))
					log.Fatalf("[dal.SetLatestSingalStatus] err %v", err)
				}
			}
		}
		err = dal.SetLatestSingalStatus(coinName, int(constdef.OpenTradeOrder))
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌开多：设置状态机状态失败 err:%v", err))
			return
		}
		utils.SendLarkNoticeToDefaultUser(ctx, "✅开多：借款成功")
	case constdef.OpenTradeOrder:
		log.Println("[OpenTradeOrder] OpenTradeOrder status.")
		//挂单
		if err := zbTradeAPI.GetCrossAsset(); err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌开多：获取杠杆账户失败 err:%v", err))
			return
		}
		tradePairInfo, err := dal.GetTradePairOrderInfo(singalKey)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌开多：获取交易信息失败 err:%v", err))
			log.Fatalf("[OpenTradeOrder.GetTradePairOrderInfo]. err. err: %v", err)
			return
		}

		if tradePairInfo.NowTradeId == "" {
			//TODO:如果redis没有存储，那也要去最近挂单历史找出有没有挂单的内容, 有可能中间直接停掉了
			//全仓交易
			qcAviableNum := zbTradeAPI.accountInfo.CrossAsset[coinName].MoneyTotalNum
			if qcAviableNum > constdef.MinQcUnit {
				log.Printf("[OpenOrder]. qcAviableNum: %v", qcAviableNum)
				id, orderPrice, orderNum, err := zbTradeAPI.SetOrder(int(constdef.Buy), coinName, qcAviableNum, int(constdef.Futures))
				utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅开多：订单号:%v, 交易币种%v,交易数量%vQC,买入价格%v", id, coinName, qcAviableNum, zbTradeAPI.RealTimeData.GetDepthData().Buys[0][0]))
				if err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, "❌开多：设置买入失败")
					log.Fatalf("[OpenOrder.SetOrder]. err: %v", err.Error())
					return
				}
				log.Printf("ku[OpenOrder]. id: %v, orderPirce: %v, orderNum: %v, err: %v\n", id, orderPrice, orderNum, err)
				if err := dal.SetNewOrder(singalKey, id, orderPrice, orderNum, int(constdef.Buy)); err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, "❌开多：存储买入交易信息失败")
					log.Fatalf("[OpenOrder.SetNewOrder]. err: %v", err.Error())
					return
				}
				//提前返回 等待一个tick检查订单状态
				return
			} else {
				err = dal.SetLatestOpertionKey(coinName, int(constdef.Buy))
				if err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, "❌开多：设置最近操作失败")
					return
				}
				err = dal.SetLatestSingalStatus(coinName, int(constdef.OpenTradeOrderFinish))
				if err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, "❌开多：设置挂单成功状态失败")
					return
				}
				utils.SendLarkNoticeToDefaultUser(ctx, "✅开多单成功，但是没有成交数量")
				return
			}
		}

		//检查订单状态，如果超过阈值，直接就取消订单
		//查询状态
		itemStatus, err := zbTradeAPI.QueryOrderStatus(constdef.CrossMarketMap[coinName], tradePairInfo.NowTradeId)
		if err != nil {
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
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅开多：买入成交信息, 币种%s, 买入价格%f, 下单个数%f, 成交个数%f, 是否成功%v", coinName, orderItem.OrderPrice, orderItem.OrderAmount, orderItem.DealAmount, itemStatus.Status == constdef.Success))
		err = dal.SetOrderInfo(singalKey, tradePairInfo.NowTradeId, orderItem)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌开多：存储买入交易信息失败")
			return
		}
		if itemStatus.Status == constdef.Success {
			err = dal.RemoveOrderInfo(singalKey)
			if err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, "❌开多：删除交易对信息失败")
				return
			}
			err = dal.SetLatestOpertionKey(coinName, int(constdef.Buy))
			if err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, "❌开多：设置最近操作失败")
				return
			}
			err = dal.SetLatestSingalStatus(coinName, int(constdef.OpenTradeOrderFinish))
			if err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, "❌开多：设置挂单成功状态失败")
				return
			}
			utils.SendLarkNoticeToDefaultUser(ctx, "✅开多单成功")
			BroadAsset(zbTradeAPI)
		} else {
			if time.Now().Unix()-tradePairInfo.OrderMap[tradePairInfo.NowTradeId].StartOrderTimeStamp > constdef.SetOrderWaitSecond {
				//取消订单
				utils.SendLarkNoticeToDefaultUser(ctx, "✅开多：挂单超时未完成，重新设置")
				if err := zbTradeAPI.CancelOrder(constdef.CrossMarketMap[coinName], tradePairInfo.NowTradeId); err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, "❌开多：取消挂单失败")
					log.Fatalf("[CancelOrder]. cancelOrder error. err: %v", err)
					return
				}
				err = dal.RemoveOrderInfo(singalKey)
				if err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, "❌开多：清除订单信息失败")
					return
				}
				utils.SendLarkNoticeToDefaultUser(ctx, "✅开多：开始重新挂单")

				return
			}
		}
	case constdef.OpenTradeOrderFinish:
		log.Println("[OpenMutipleOrder] Finish.")
		return
	default:
		return
	}
}

//平多单
func CloseMutipleOrder(zbTradeAPI *ZbLive, coinName string, newSingalKey string) {
	ctx := context.Background()
	status, err := dal.GetLatestSingalStatus(coinName)
	if err != nil {
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平多：获取状态机状态失败 err %v", err))
		log.Fatalf("[CloseMutipleOrder] redis error")
		return
	}
	latestSingalKey, err := dal.GetLatestSinglKey(coinName)
	if err != nil {
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平多：获取当前交易key失败 err %v", err))
		log.Fatalf("[CloseMutipleOrder]. latestSingalKey: %v", latestSingalKey)
		return
	}
	if latestSingalKey == newSingalKey {
		return
	}
	switch status {
	case constdef.OpenTradeOrderFinish:
		log.Print("[CloseMutipleOrder] OpenTradeOrderFinish.")
		err = dal.SetLatestSingalStatus(coinName, int(constdef.CloseTradeOrder))
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平多：获取状态机状态失败 err %v", err))
			return
		}
		utils.SendLarkNoticeToDefaultUser(ctx, "✅开始平多单")

	case constdef.CloseTradeOrder:
		log.Print("[CloseMutipleOrder] CloseTradeOrder.")
		if err := zbTradeAPI.GetCrossAsset(); err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌平多：获取杠杆账号失败")
			log.Fatalf("[CloseMutipleOrder.GetCrossAsset]. err: %v", err.Error())
			return
		}
		crossAsset := zbTradeAPI.accountInfo.CrossAsset[coinName]
		sellVritualCoinNum := crossAsset.VirtualTotalNum
		var id string
		latestSignalKey, err := dal.GetLatestSinglKey(coinName)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌平多：获取交易key失败")
			log.Fatalf("[CloseMutipleOrder.GetLatestSinglKey]. err: %v", err)
			return
		}
		tradePairInfo, err := dal.GetTradePairOrderInfo(latestSignalKey)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌平多：获取交易对信息失败")
			log.Fatalf("[CloseMutipleOrder.GetTradePairOrderInfo]. err. err: %v", err)
			return
		}
		if tradePairInfo.NowTradeId == "" {
			if sellVritualCoinNum > constdef.MinCoinUnit[coinName] {
				//挂单
				var orderErr error
				var orderPrice, orderNum float64
				log.Printf("[CloseMutipleOrder] coinName: %v, sellVritualCoinNum: %v", coinName, sellVritualCoinNum)
				utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅平多：剩余可出售%v%s", sellVritualCoinNum, coinName))

				id, orderPrice, orderNum, orderErr = zbTradeAPI.SetOrder(int(constdef.Sell), coinName, sellVritualCoinNum, int(constdef.Futures))
				if orderErr != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平多：平多挂卖单失败 err %v", orderErr))
					log.Fatalf("[CloseMutipleOrder.SetOrder]. err: %v", orderErr.Error())
					return
				}
				utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅平多：订单号: %v, 挂单出售%v%s, 出售价格%f", id, orderNum, coinName, orderPrice))

				//todo 如果设置redis失败，可能会导致重复挂单
				if err := dal.SetNewOrder(latestSignalKey, id, orderPrice, orderNum, int(constdef.Sell)); err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平多：设置挂单信息失败 err %v", err))
					log.Fatalf("[CloseMutipleOrder.SetNewOrder]. err: %v", err.Error())
					return
				}
				//提前返回 等待一个tick检查订单状态
				return
			} else {
				err = dal.SetLatestSingalStatus(coinName, int(constdef.ClosePayBack))
				if err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平多：设置状态机状态失败 err %v", err))
					return
				}
				utils.SendLarkNoticeToDefaultUser(ctx, "✅平多:挂单卖出成功")
				return
			}
		}
		//检查订单状态，如果超过阈值，直接就取消订单
		//查询状态
		itemStatus, err := zbTradeAPI.QueryOrderStatus(constdef.CrossMarketMap[coinName], tradePairInfo.NowTradeId)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平多：获取卖出订单状态失败 err %v", err))
			log.Fatalf("[CloseMutipleOrder.QueryOrderStatus]. err: %v", err)
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
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅平多：挂单状态: 订单号:%v,挂单出售%f%s, 出售价格%f, 成交个数%f, 是否成功%v", tradePairInfo.NowTradeId, orderItem.OrderAmount, coinName, orderItem.OrderPrice, orderItem.DealAmount, itemStatus.Status == constdef.Success))
		log.Printf("[CloseMutipleOrder], orderItem: %v", orderItem)
		err = dal.SetOrderInfo(latestSignalKey, tradePairInfo.NowTradeId, orderItem)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平多：设置订单信息失败 err %v", err))
			return
		}
		if itemStatus.Status == constdef.Success {
			err = dal.RemoveOrderInfo(latestSignalKey)
			if err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平多：清除当前订单失败 err %v", err))
				return
			}
			err = dal.SetLatestSingalStatus(coinName, int(constdef.ClosePayBack))
			if err != nil {
				utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平多：设置状态机状态失败 err %v", err))
				return
			}
			utils.SendLarkNoticeToDefaultUser(ctx, "✅平多:卖出委托单成交成功")

		} else {
			if time.Now().Unix()-tradePairInfo.OrderMap[tradePairInfo.NowTradeId].StartOrderTimeStamp > constdef.SetOrderWaitSecond {
				//取消订单
				utils.SendLarkNoticeToDefaultUser(ctx, "✅平多：挂单超时未完成，重新设置")
				if err := zbTradeAPI.CancelOrder(constdef.CrossMarketMap[coinName], tradePairInfo.NowTradeId); err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, "❌平多：取消挂单失败")
					log.Fatalf("[CancelOrder]. cancelOrder error. err: %v", err)
					return
				}
				err = dal.RemoveOrderInfo(latestSingalKey)
				if err != nil {
					utils.SendLarkNoticeToDefaultUser(ctx, "❌平多：清除订单信息失败")
				}
				utils.SendLarkNoticeToDefaultUser(ctx, "✅平多：开始重新挂单")
				return
			}
		}
	case constdef.ClosePayBack:
		log.Print("[CloseMutipleOrder] ClosePayBack.")
		// 记录一下还款币种，还款个数
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅平多：准备还款QC"))
		latestSignalKey, err := dal.GetLatestSinglKey(coinName)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌平多：获取交易key失败")
			log.Fatalf("[CloseEmptyOrder.GetLatestSinglKey]. err: %v", err)
			return
		}
		err = dal.SetLoanInfo(latestSignalKey, "QC", zbTradeAPI.accountInfo.CrossAsset[coinName].MoneyFeed)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌平多：设置借款利息失败")
			log.Fatalf("[SetLoanInfo]. err: %v", err)
			return
		}
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅平多：存储借款利息成功, 币种:QC, 个数%v", zbTradeAPI.accountInfo.CrossAsset[coinName].MoneyFeed))
		needPayBack := zbTradeAPI.accountInfo.CrossAsset[coinName].MoneyLoadIn
		//还款 还法币
		err = zbTradeAPI.DoCrossRepay("QC", constdef.CrossMarketMap[coinName])
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌平多：还款失败")
			log.Fatalf("[CloseMutipleOrder.DoCrossRepay]. err: %v", err)
			return
		}
		utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("✅平多：还款成功, 还款%vQC", needPayBack))
		err = dal.SetLatestSingalStatus(coinName, int(constdef.CloseTradeOrderFinish))
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌平多：设置状态机状态失败")
			return
		}
		utils.SendLarkNoticeToDefaultUser(ctx, "✅平多:还款状态完成")
	case constdef.CloseTradeOrderFinish:
		log.Print("[CloseMutipleOrder] CloseTradeOrderFinish.")
		err = dal.SetLatestSingalStatus(coinName, int(constdef.TradeOrderEndHandler))
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌设置状态机状态失败")
		}
		utils.SendLarkNoticeToDefaultUser(ctx, "✅平多：处理平仓完成")

	case constdef.TradeOrderEndHandler:
		log.Print("[CloseMutipleOrder] TradeOrderEndHandler.")
		utils.SendLarkNoticeToDefaultUser(ctx, "✅平多：开始计算收益")
		err := CalculateTradeStatus(zbTradeAPI, constdef.Buy, coinName)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平多：计算收益失败 err %v", err))
			log.Fatalf("[CalculateTradeStatus] err %v", err)
			return
		}
		err = dal.DelLatestOperation(coinName)
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, fmt.Sprintf("❌平多：清理最近一个操作错误。err:%v", err))
			log.Fatalf("[CloseMutipleOrder.DelLatestOperation]. err: %v", err)
			return
		}
		err = dal.SetLatestSingalStatus(coinName, int(constdef.TradeOrderFinish))
		if err != nil {
			utils.SendLarkNoticeToDefaultUser(ctx, "❌平多：设置状态机状态失败")
			return
		}
		utils.SendLarkNoticeToDefaultUser(ctx, "✅平多：计算利润完成")
		BroadAsset(zbTradeAPI)

	case constdef.TradeOrderFinish:
		log.Print("[CloseMutipleOrder] TradeOrderFinish.")
		utils.SendLarkNoticeToDefaultUser(ctx, "✅平多结束")
		return
	default:
		log.Printf("[CloseMutipleOrder]: %v", status)
		return
	}
}
