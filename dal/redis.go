package dal

import (
	"black_hole/constdef"
	"black_hole/model"
	"black_hole/utils"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"time"
)

var (
	OrderStorage *redis.Client
)

func init() {
	OrderStorage = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

//获取订单信息： 买入、卖出
func GetTradePairOrderInfo(key string) (model.TradePair, error) {
	var result model.TradePair
	val, err := OrderStorage.HGetAll(key).Result()
	if err != nil && redis.Nil != err {
		log.Fatalf("[getTradePairOrderInfo]. err: %v", err)
		return result, err
	}
	if redis.Nil == err {
		return result, nil
	}
	result = model.TradePair{
		NowTradeId:    val["NowTradeId"],
		ProfitPercent: utils.StrToFloat64(val["ProfitPercent"]),
	}
	result.OrderMap = make(map[string]model.OrderItem)
	//还原每一个交易对
	for k, v := range val {
		if constdef.ContainOrderDetailKey(k) {
			var orderDetailMap model.OrderItem
			err := json.Unmarshal([]byte(v), &orderDetailMap)
			if err != nil {
				return result, err
			}
			result.OrderMap[constdef.GetOrderIdByKey(k)] = orderDetailMap
		}
	}
	return result, nil
}

//获取最近状态
func GetLatestSingalStatus(coinName string) (constdef.TradeOrderStatus, error) {
	val, err := OrderStorage.HGet(fmt.Sprintf(constdef.LatestKey, coinName), constdef.LatestSingalStatusKey).Int64()
	if err != nil && redis.Nil != err {
		log.Fatalf("[GetLatestSingalStatus]. err: %v", err)
		return constdef.TradeOrderStatus(val), err
	}
	log.Printf("[GetLatestSingalStatus]: val: %v, err: %v", val, err)
	if redis.Nil == err {
		return constdef.OpenOrderUninit, nil
	}
	return constdef.TradeOrderStatus(val), err
}

func SetLatestSingalStatus(coinName string, status int) error {
	log.Printf("[SetLatestSingalStatus] status: %v", status)
	_, err := OrderStorage.HSet(fmt.Sprintf(constdef.LatestKey, coinName), constdef.LatestSingalStatusKey, status).Result()
	return err
}

func DelLatestSingalStatus(coinName string) error {
	log.Printf("[DelLatestSingalStatus]")
	_, err := OrderStorage.HDel(fmt.Sprintf(constdef.LatestKey, coinName), constdef.LatestSingalStatusKey).Result()
	return err
}

func GetLatestSinglKey(coinName string) (string, error) {
	val, err := OrderStorage.HGet(fmt.Sprintf(constdef.LatestKey, coinName), constdef.LatestSingalKey).Result()
	if err != nil && err != redis.Nil {
		return val, err
	}
	return val, nil
}

func SetLatestSingalKey(coinName string, key string) error {
	_, err := OrderStorage.HSet(fmt.Sprintf(constdef.LatestKey, coinName), constdef.LatestSingalKey, key).Result()
	return err
}

func SetNewOrder(key string, orderId string, orderPirce float64, assetNum float64, TradeType int) error {
	args := make(map[string]interface{})
	args["NowTradeId"] = orderId
	orderItem := &model.OrderItem{
		TradeStatus:         int(constdef.Processing),
		OrderPrice:          orderPirce,
		OrderAmount:         assetNum,
		DealAmount:          0,
		StartOrderTimeStamp: time.Now().Unix(),
		EndOrderTimeStamp:   0,
		SellOrBuy:           int(constdef.Buy),
	}
	orderItemStructByte, err := json.Marshal(orderItem)
	if err != nil {
		log.Fatalf("[SetNewOrder]. err: %v", err)
		return err
	}
	args[fmt.Sprintf(constdef.OrderDetailKey, orderId)] = orderItemStructByte
	_, err = OrderStorage.HMSet(key, args).Result()
	return err
}

//写入状态
func SetOrderInfo(key string, orderId string, item model.OrderItem) error {
	args := make(map[string]interface{})
	orderItemStructByte, err := json.Marshal(item)
	if err != nil{
		log.Fatalf("[SetOrderInfo] err %v", err)
		return err
	}
	args[fmt.Sprintf(constdef.OrderDetailKey, orderId)] = orderItemStructByte
	_, err = OrderStorage.HMSet(key, args).Result()
	return err
}

//擦掉orderId
func RemoveOrderInfo(key string) error {
	args := make(map[string]interface{})
	args["NowTradeId"] = ""
	_, err := OrderStorage.HMSet(key, args).Result()
	return err
}

func SetLoanInfo(key, coinName string, loanNum float64) error {
	args := make(map[string]interface{})
	loanInfo := &model.LoanInfo{
		LoanCoin: coinName,
		LoanNum: loanNum,
	}
	info , _ := json.Marshal(loanInfo)
	args[constdef.LoanInfo] = info
	_, err := OrderStorage.HMSet(key, args).Result()
	return err
}

func GetLoanInfo(key string) (*model.LoanInfo, error){
	infoModel := &model.LoanInfo{}
	val, err := OrderStorage.HGetAll(key).Result()
	if err != nil && redis.Nil != err {
		log.Fatalf("[getTradePairOrderInfo]. err: %v", err)
		return nil, err
	}
	if redis.Nil == err {
		return infoModel, nil
	}
	info := val[constdef.LoanInfo]
	err = json.Unmarshal([]byte(info), infoModel)
	if err != nil{
		return nil, err
	}
	return infoModel, nil
}

func SetIncomePercent(key string, incomePercent float64) error {
	args := make(map[string]interface{})
	args["ProfitPercent"] = incomePercent
	_, err := OrderStorage.HMSet(key, args).Result()
	return err
}

func SetLatestOpertionKey(coinName string, operation int) error {
	_, err := OrderStorage.HSet(fmt.Sprintf(constdef.LatestKey, coinName), constdef.LatestOperationKey, operation).Result()
	return err
}

func GetLatestOpertion(coinName string) (constdef.TradeType, error) {
	val, err := OrderStorage.HGet(fmt.Sprintf(constdef.LatestKey, coinName), constdef.LatestOperationKey).Int64()
	if err == redis.Nil {
		return constdef.StandBy, nil
	}
	if err != nil {
		return constdef.StandBy, err
	}
	return constdef.TradeType(val), nil
}

func DelLatestOperation(coinName string) error {
	_, err := OrderStorage.HDel(fmt.Sprintf(constdef.LatestKey, coinName), constdef.LatestOperationKey).Result()
	return err
}