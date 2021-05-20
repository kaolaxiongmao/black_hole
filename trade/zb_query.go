package trade

import (
	"black_hole/constdef"
	"black_hole/model"
	"black_hole/utils"
	"errors"
	"github.com/bitly/go-simplejson"
	"log"
	"strconv"
)

//主要是一些查询接口 目前是查询下单是否下完
//查询订单的状态
func (zb *ZbLive) QueryOrderStatus(market string, id string) (model.OrderItemStatus, error) {
	path := constdef.TradeZbUrl + "/getOrder?"
	paramsPublicMap := make(map[string]string)
	paramsPublicMap["accesskey"] = zb.accessKey
	paramsPublicMap["method"] = "getOrder"
	paramsPublicMap["currency"] = market
	paramsPublicMap["id"] = id
	publicParams := utils.JoinStringsInASCII(paramsPublicMap)
	enSercetKey := utils.Sha1Algro(zb.sercetKey)
	var orderItemStatus model.OrderItemStatus
	//按照ascii排序
	sign := utils.Hmac(enSercetKey, publicParams)
	reqTimeParams := "&reqTime=" + strconv.FormatInt(utils.GetReqTime(), 10)
	url := path + publicParams + "&sign=" + sign + reqTimeParams
	requst, err := utils.HttpGet(url)
	if err != nil {
		return orderItemStatus, err
	}
	jsonStruct, _ := simplejson.NewJson([]byte(requst))
	jsonStr, err := jsonStruct.EncodePretty()
	log.Print("[QueryOrderStatus]: %v, %v", string(jsonStr), err)
	_, err = jsonStruct.Get("code").Int()
	if err != nil {
		//如果没有code字段就表示成功了
		total_amount, err := jsonStruct.Get("total_amount").Float64()
		if err != nil {
			log.Fatalf("[QueryOrderStatus]. format err :%v", err)
			return orderItemStatus, err
		}
		trade_amount, err := jsonStruct.Get("trade_amount").Float64()
		if err != nil {
			log.Fatalf("[QueryOrderStatus]. format err :%v", err)
			return orderItemStatus, err
		}
		status, err := jsonStruct.Get("status").Int()
		if err != nil {
			log.Fatalf("[QueryOrderStatus]. format err :%v", err)
			return orderItemStatus, err
		}
		orderType, err := jsonStruct.Get("type").Int()
		if err != nil {
			log.Fatalf("[QueryOrderStatus]. format err :%v", err)
			return orderItemStatus, err
		}
		price, err := jsonStruct.Get("price").Float64()
		if err != nil {
			log.Fatalf("[QueryOrderStatus]. format err :%v", err)
			return orderItemStatus, err
		}
		tradeDate, err := jsonStruct.Get("trade_date").Int64()
		if err != nil {
			log.Fatalf("[QueryOrderStatus]. format err :%v", err)
			return orderItemStatus, err
		}
		orderItemStatus.Price = price
		orderItemStatus.DealAmount = trade_amount
		orderItemStatus.OrderAmount = total_amount
		orderItemStatus.Type = orderType
		//毫秒时间戳 -> 秒时间戳
		orderItemStatus.Date = tradeDate/1000
		if total_amount == trade_amount {
			//数量相等 就认为成功
			orderItemStatus.Status = constdef.Success
			log.Print("[QueryOrderStatus]. order success")
			return orderItemStatus, nil
		} else {
			if status == int(constdef.Canceled) {
				orderItemStatus.Status = constdef.Canceled
				log.Print("[QueryOrderStatus]. order cacnel")
				return orderItemStatus, nil
			} else {
				log.Print("[QueryOrderStatus]. order processing")
				orderItemStatus.Status = constdef.Processing
				return orderItemStatus, nil
			}
		}
	}
	message, err := jsonStruct.Get("message").String()
	if err == nil {
		log.Fatalf("[QueryOrderStatus]: code not 1000. message:  %v", message)
		return orderItemStatus, errors.New(message)
	} else {
		log.Fatalf("[QueryOrderStatus]: message not parse")
		return orderItemStatus, err
	}
}
