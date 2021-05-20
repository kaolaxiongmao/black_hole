package trade

import (
	"black_hole/conf"
	"black_hole/constdef"
	"black_hole/data"
	"black_hole/model"
	"black_hole/utils"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"log"
	"strconv"
	"time"
)

//zb.live相关
type ZbLive struct {
	accessKey       string
	sercetKey       string
	passWord        string
	focusCoins      []string          //关注的币信息
	accountInfo     model.AccountInfo //账户资产信息 现货
	RealTimeData    *data.ZbLiveData  //这里可以拿到事实数据
	crossTradePairs []string          //期货交易对
	mutipleLever    float64           //做多杠杆率
	emptyLever      float64           //做空杠杆率
	barTime         string
	currencyName    string
}

//tradeType是用来区分高杠杆和低倍杆杆 进而分为2个账户
func NewTradeAPI() *ZbLive {
	var zbTradeAPI *ZbLive
	config := conf.GetConfig()
	zbTradeAPI = &ZbLive{
		accessKey:       config.LowAccessKey,
		sercetKey:       config.LowSecretKey,
		passWord:        config.LowPwd,
		focusCoins:      constdef.FocusCoins,
		crossTradePairs: constdef.LeverCrossCoins,
		mutipleLever:    config.MultipleLever,
		emptyLever:      config.EmptyLever,
		barTime:         config.BarTime,
		currencyName:    config.CurCurrency,
	}
	BroadInit(zbTradeAPI)
	zbTradeAPI.accountInfo.CrossAsset = make(map[string]model.LeverAccountInfo)
	zbTradeAPI.accountInfo.VirtualCurrency = make(map[string]float64)
	return zbTradeAPI
}

//获取现货和期货仓位
func (zb *ZbLive) GetAccountInfo() error {
	path := constdef.TradeZbUrl + "/getAccountInfo?"
	paramsPublicMap := make(map[string]string)
	paramsPublicMap["accesskey"] = zb.accessKey
	paramsPublicMap["method"] = "getAccountInfo"
	publicParams := utils.JoinStringsInASCII(paramsPublicMap)
	enSercetKey := utils.Sha1Algro(zb.sercetKey)
	//按照ascii排序
	sign := utils.Hmac(enSercetKey, publicParams)
	reqTimeParams := "&reqTime=" + strconv.FormatInt(utils.GetReqTime(), 10)
	url := path + publicParams + "&sign=" + sign + reqTimeParams
	log.Printf("[GetAccountInfo]. url: %v", url)
	requst, err := utils.HttpGet(url)
	if err != nil {
		log.Fatalf("[GetAccountInfo]. err: %v", err)
		return err
	}
	jsonStruct, _ := simplejson.NewJson([]byte(requst))
	//jsonStr, err := jsonStruct.EncodePretty()
	//log.Print("[GetAccountInfo]: %v, %v", string(jsonStr), err)
	code, err := jsonStruct.Get("code").Int()
	if err == nil && code != 1000 {
		message, err := jsonStruct.Get("message").String()
		if err == nil {
			log.Fatalf("[GetAccountInfo]: code not 1000. message:  %v", message)
			return errors.New(message)
		} else {
			log.Fatalf("[GetAccountInfo]: message not parse")
			return err
		}
	}
	coinArray, _ := jsonStruct.Get("result").Get("coins").Array()
	for _, coinInfo := range coinArray {
		for _, coinName := range zb.focusCoins {
			if coinName == coinInfo.(map[string]interface{})["cnName"].(string) {
				value, _ := strconv.ParseFloat(coinInfo.(map[string]interface{})["available"].(string), 10)
				zb.accountInfo.VirtualCurrency[coinName] = value
			}
		}
	}
	if err := zb.GetCrossAsset(); err != nil {
		log.Fatalf("[GetAccountInfo], get crossAsset failed. err: %v", err.Error())
		return err
	}
	return nil
}

//单个单子: 设置订单
//买入 assetValue: 表示是虚拟币数量
//卖出 assetValue: 也表示卖出多少个虚拟货币
//返回：订单号、挂单价格、挂单数量、
func (zb *ZbLive) SetOrder(tradeType int, coinName string, assetValue float64, acctType int) (string, float64, float64, error) {
	path := constdef.TradeZbUrl + "/order?"
	paramsPublicMap := make(map[string]string)
	paramsPublicMap["accesskey"] = zb.accessKey
	paramsPublicMap["method"] = "order"
	paramsPublicMap["currency"] = constdef.MarketMap[coinName]
	paramsPublicMap["tradeType"] = strconv.Itoa(tradeType)
	paramsPublicMap["acctType"] = strconv.Itoa(acctType)
	var orderPrice float64
	var orderId string
	//深度数据大与5秒则不处理 [远程服务器时间和本地时间]
	if float64(time.Now().Unix()-zb.RealTimeData.GetDepthData().LocalTimeStamp) > 5 {
		log.Fatalf("[SetOrder]: err: TimeStamp not fresh: now: %v, data_timeStamp: %v", time.Now().Unix(), zb.RealTimeData.GetDepthData().TimeStamp)
		return orderId, 0, 0, errors.New("depth time not fresh")
	}

	//买入策略
	if tradeType == int(constdef.Buy) {
		//卖一 卖二 以此类推 算出钱可以买多少个 击穿买入方法
		orderPrice = zb.RealTimeData.GetDepthData().Buys[0][0]
		assetValue = assetValue / orderPrice
	} else {
		//卖单： 买1 买2 有多少资产可以挂多少个
		orderPrice = zb.RealTimeData.GetDepthData().Sells[0][0]
	}
	paramsPublicMap["price"] = fmt.Sprintf("%v", orderPrice)
	paramsPublicMap["amount"] = fmt.Sprintf("%v", assetValue)
	publicParams := utils.JoinStringsInASCII(paramsPublicMap)
	enSercetKey := utils.Sha1Algro(zb.sercetKey)
	//按照ascii排序
	sign := utils.Hmac(enSercetKey, publicParams)
	reqTimeParams := "&reqTime=" + strconv.FormatInt(utils.GetReqTime(), 10)
	url := path + publicParams + "&sign=" + sign + reqTimeParams
	log.Printf("[SetOrder]: url: %s, %v", url, orderPrice)
	requst, err := utils.HttpGetWithTimeout(url)
	log.Printf("[SetOrder]: %s", requst)
	//json struct
	if err != nil {
		log.Fatalf("[SetOrder.HttpGet]: err: %s", err.Error())
		return orderId, 0, 0, err
	}
	jsonStruct, err := simplejson.NewJson([]byte(requst))
	if err != nil {
		log.Fatalf("[SetOrder.NewJson]: err: %s request: %s", err.Error(), requst)
		return orderId, 0, 0, err
	}
	jsonStr, err := jsonStruct.EncodePretty()
	log.Printf("[SetOrder]: %v, %v", string(jsonStr), err)
	code, err := jsonStruct.Get("code").Int()
	if err != nil {
		log.Fatalf("[SetOrder]: code error.")
		return orderId, 0, 0, err
	}
	if code != 1000 {
		message, err := jsonStruct.Get("message").String()
		if err == nil {
			log.Fatalf("[SetOrder]: code not 1000. message:  %v", message)
			return orderId, 0, 0, errors.New(message)
		} else {
			log.Fatalf("[SetOrder]: message not parse")
			return orderId, 0, 0, err
		}
	} else {
		orderId, err = jsonStruct.Get("id").String()
		if err != nil {
			log.Fatalf("[SetOrder]: id error. err: %v", err)
			return orderId, 0, 0, err
		}
		log.Print("[SetOrder]. orderId: %v", orderId)
		return orderId, orderPrice, assetValue, nil
	}
}

//取消订单
func (zb *ZbLive) CancelOrder(market string, orderId string) error {
	path := constdef.TradeZbUrl + "/cancelOrder?"
	paramsPublicMap := make(map[string]string)
	paramsPublicMap["accesskey"] = zb.accessKey
	paramsPublicMap["method"] = "cancelOrder"
	paramsPublicMap["currency"] = market
	paramsPublicMap["id"] = orderId
	publicParams := utils.JoinStringsInASCII(paramsPublicMap)
	enSercetKey := utils.Sha1Algro(zb.sercetKey)
	//按照ascii排序
	sign := utils.Hmac(enSercetKey, publicParams)
	reqTimeParams := "&reqTime=" + strconv.FormatInt(utils.GetReqTime(), 10)
	url := path + publicParams + "&sign=" + sign + reqTimeParams
	log.Printf("[CancelOrder]. url: %v", url)
	requst, err := utils.HttpGet(url)
	if err != nil {
		return err
	}
	jsonStruct, _ := simplejson.NewJson([]byte(requst))
	jsonStr, err := jsonStruct.EncodePretty()
	log.Printf("[CancelOrder]: %v, %v", string(jsonStr), err)
	code, err := jsonStruct.Get("code").Int()
	if err != nil {
		log.Fatalf("[CancelOrder]: code error.")
		return err
	}
	if code != 1000 && code != 3001 {
		message, err := jsonStruct.Get("message").String()
		if err == nil {
			log.Fatalf("[CancelOrder]: code not 1000. message:  %v", message)
			return errors.New(message)
		} else {
			log.Fatalf("[CancelOrder]: message not parse")
			return err
		}
	}
	log.Printf("[CancelOrder]. cancel success. market: %v, orderId: %v", market, orderId)
	return nil
}

func (zb *ZbLive) GetCurAccountInfo() model.AccountInfo {
	return zb.accountInfo
}

func SetLever(config conf.Config, zbtradeAPI *ZbLive) *ZbLive {
	zbtradeAPI.emptyLever = config.EmptyLever
	zbtradeAPI.mutipleLever = config.MultipleLever
	return zbtradeAPI
}
