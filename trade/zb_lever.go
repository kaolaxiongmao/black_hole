package trade

import (
	"black_hole/constdef"
	"black_hole/model"
	"black_hole/utils"
	"encoding/json"
	"errors"
	"github.com/bitly/go-simplejson"
	"log"
	"strconv"
)

//中币杠杆交易接口
//包括转账到杠杆账户、把杠杆账户的钱转账到现货系统
//获取全仓信息
func (zb *ZbLive) GetCrossAsset() error {
	path := constdef.TradeZbUrl + "/getLeverAssetsInfo?"
	paramsPublicMap := make(map[string]string)
	paramsPublicMap["accesskey"] = zb.accessKey
	paramsPublicMap["method"] = "getLeverAssetsInfo"
	publicParams := utils.JoinStringsInASCII(paramsPublicMap)
	enSercetKey := utils.Sha1Algro(zb.sercetKey)
	sign := utils.Hmac(enSercetKey, publicParams)
	reqTimeParams := "&reqTime=" + strconv.FormatInt(utils.GetReqTime(), 10)
	url := path + publicParams + "&sign=" + sign + reqTimeParams
	//log.Printf("[]: url: %s", url)
	requst, err := utils.HttpGetWithTimeout(url)
	//log.Printf("[GetCrossAsset]: request: %s, err: %v", requst, err)
	//json struct
	if err != nil {
		log.Fatalf("[GetCrossAsset.HttpGet]: err: %s", err.Error())
		return err
	}
	jsonStruct, err := simplejson.NewJson([]byte(requst))
	if err != nil {
		log.Fatalf("[GetCrossAsset.NewJson]: err: %s", err.Error())
		return err
	}
	//jsonStr, err := jsonStruct.EncodePretty()
	//log.Print("[GetCrossAsset.EncodePretty]: %v, %v", string(jsonStr), err)
	code, err := jsonStruct.Get("code").Int()
	if err != nil {
		log.Fatalf("[GetCrossAsset]: code error.")
		return err
	}

	if code != 1000 {
		message, err := jsonStruct.Get("message").String()
		if err == nil {
			log.Fatalf("[GetCrossAsset]: code not 1000. message:  %v", message)
			return errors.New(message)
		} else {
			log.Fatalf("[GetCrossAsset]: message not parse")
			return err
		}
	}
	levers, err := jsonStruct.Get("message").Get("datas").Get("levers").Array()
	if err != nil {
		log.Fatalf("[GetCrossAsset]. levers err: %v", err)
		return err
	}

	// "fAvailable": 当前提供的法币数量。
	// （包含杆杆） "fLoanIn":  法币接入的数量。
	//"fOverdraft": 法币货币利息
	// "cAvailable": 数字货币账户数量
	// "cLoanIn": 数字货币借的数量
	//"cOverdraft": 数字货币利息
	//"key": 当前的交易对 格式：btcqc  ltcqc
	for _, pair := range zb.crossTradePairs {
		for _, lever := range levers {
			leverStruct := &model.LeverStrAccountInfo{}
			leverByte, err := json.Marshal(lever)
			if err != nil {
				log.Fatalf("[GetCrossAsset]. Marshal, err: %v", err)
				return err
			}
			err = json.Unmarshal(leverByte, leverStruct)
			if err != nil {
				log.Fatalf("[GetCrossAsset]. Unmarshal, err: %v", err)
				return err
			}
			if leverStruct.TradePair == pair {
				leverNumStruct := model.LeverAccountInfo{}
				VirtualNum, err := strconv.ParseFloat(leverStruct.CAvailable, 10)
				if err != nil {
					log.Fatalf("[GetCrossAsset]. ParseFloat, err: %v", err)
					return err
				}
				VirtualLoadIn, err := strconv.ParseFloat(leverStruct.CLoanIn, 10)
				if err != nil {
					log.Fatalf("[GetCrossAsset]. ParseFloat, err: %v", err)
					return err
				}
				virtualFeed, err := strconv.ParseFloat(leverStruct.COverdraft, 10)
				if err != nil {
					log.Fatalf("[GetCrossAsset]. ParseFloat, err: %v", err)
					return err
				}
				MoneyTotalNum, err := strconv.ParseFloat(leverStruct.FAvailable, 10)
				if err != nil {
					log.Fatalf("[GetCrossAsset]. ParseFloat, err: %v", err)
					return err
				}
				MoneyLoadIn, err := strconv.ParseFloat(leverStruct.FLoanIn, 10)
				if err != nil {
					log.Fatalf("[GetCrossAsset]. ParseFloat, err: %v", err)
					return err
				}
				moneyFeed, err := strconv.ParseFloat(leverStruct.FOverDraft, 10)
				if err != nil {
					log.Fatalf("[GetCrossAsset]. ParseFloat, err: %v", err)
					return err
				}
				unwindPrice, err := strconv.ParseFloat(leverStruct.UnwindPrice, 10)
				if err != nil {
					log.Fatalf("[GetCrossAsset]. ParseFloat, err: %v", err)
					return err
				}
				netQC, err := strconv.ParseFloat(leverStruct.NetQC, 10)
				if err != nil {
					log.Fatalf("[GetCrossAsset]. ParseFloat, err: %v", err)
					return err
				}
				leverNumStruct.VirtualTotalNum = VirtualNum
				leverNumStruct.VirtualLoadIn = VirtualLoadIn
				leverNumStruct.VirtualFeed = virtualFeed
				leverNumStruct.NetQC = netQC
				leverNumStruct.MoneyTotalNum = MoneyTotalNum
				leverNumStruct.MoneyLoadIn = MoneyLoadIn
				leverNumStruct.MoneyFeed = moneyFeed
				leverNumStruct.UnwindPrice = unwindPrice
				leverNumStruct.RepayLeverShow = leverStruct.RepayLeverShow

				//BTC/LTC/
				zb.accountInfo.CrossAsset[leverStruct.CoinName] = leverNumStruct
			}
		}
	}
	log.Print("[GetCrossAsset]: %v", zb.accountInfo.CrossAsset)
	return nil
}

//借款
func (zb *ZbLive) DoCrossLoan(amount float64, coinName string, marketName string) error {
	path := constdef.TradeZbUrl + "/autoBorrow?"
	paramsPublicMap := make(map[string]string)
	paramsPublicMap["accesskey"] = zb.accessKey
	paramsPublicMap["method"] = "autoBorrow"
	paramsPublicMap["coin"] = coinName
	paramsPublicMap["marketName"] = marketName
	paramsPublicMap["amount"] = strconv.FormatFloat(amount, 'f', -1, 64)
	paramsPublicMap["safePwd"] = zb.passWord
	publicParams := utils.JoinStringsInASCII(paramsPublicMap)
	enSercetKey := utils.Sha1Algro(zb.sercetKey)
	sign := utils.Hmac(enSercetKey, publicParams)
	reqTimeParams := "&reqTime=" + strconv.FormatInt(utils.GetReqTime(), 10)
	url := path + publicParams + "&sign=" + sign + reqTimeParams
	log.Printf("[doCrossLoan]: url: %s", url)
	requst, err := utils.HttpGetWithTimeout(url)
	log.Printf("[doCrossLoan]: request: %s, err: %v", requst, err)
	//json struct
	if err != nil {
		log.Fatalf("[doCrossLoan.HttpGet]: err: %s", err.Error())
		return err
	}
	jsonStruct, err := simplejson.NewJson([]byte(requst))
	if err != nil {
		log.Fatalf("[doCrossLoan.NewJson]: err: %s request: %s", err.Error(), requst)
		return err
	}
	jsonStr, err := jsonStruct.EncodePretty()
	log.Printf("[doCrossLoan]: %v, %v", string(jsonStr), err)
	code, err := jsonStruct.Get("code").Int()
	if err != nil {
		log.Fatalf("[doCrossLoan]: code error.")
		return err
	}
	if code != 1000 {
		message, err := jsonStruct.Get("message").String()
		if err == nil {
			log.Fatalf("[doCrossLoan]: code not 1000. message:  %v", message)
			return errors.New(message)
		} else {
			log.Fatalf("[doCrossLoan]: message not parse")
			return err
		}
	}
	return nil
}

//默认每次还款都是全部还款，不能保留部分资金在杠杆中 防止增加风险
func (zb *ZbLive) DoCrossRepay(coinName string, marketName string) error {
	path := constdef.TradeZbUrl + "/doAllRepay?"
	paramsPublicMap := make(map[string]string)
	paramsPublicMap["accesskey"] = zb.accessKey
	paramsPublicMap["method"] = "doAllRepay"
	paramsPublicMap["coin"] = coinName
	paramsPublicMap["marketName"] = marketName
	publicParams := utils.JoinStringsInASCII(paramsPublicMap)
	enSercetKey := utils.Sha1Algro(zb.sercetKey)
	sign := utils.Hmac(enSercetKey, publicParams)
	reqTimeParams := "&reqTime=" + strconv.FormatInt(utils.GetReqTime(), 10)
	url := path + publicParams + "&sign=" + sign + reqTimeParams
	log.Printf("[doCrossRepay]: url: %s", url)
	requst, err := utils.HttpGetWithTimeout(url)
	log.Printf("[doCrossRepay]: request: %s, err: %v", requst, err)
	//json struct
	if err != nil {
		log.Fatalf("[doCrossRepay.HttpGet]: err: %s", err.Error())
		return err
	}
	jsonStruct, err := simplejson.NewJson([]byte(requst))
	if err != nil {
		log.Fatalf("[doCrossRepay.NewJson]: err: %s request: %s", err.Error(), requst)
		return err
	}
	jsonStr, err := jsonStruct.EncodePretty()
	log.Print("[doCrossRepay]: %v, %v", string(jsonStr), err)
	code, err := jsonStruct.Get("code").Int()
	if err != nil {
		log.Fatalf("[doCrossRepay]: code error.")
		return err
	}
	if code != 1000 {
		message, err := jsonStruct.Get("message").String()
		if err == nil {
			log.Fatalf("[doCrossRepay]: code not 1000. message:  %v", message)
			return errors.New(message)
		} else {
			log.Fatalf("[doCrossRepay]: message not parse")
			return err
		}
	}
	return nil
}
