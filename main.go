package main

import (
	"black_hole/conf"
	"black_hole/constdef"
	"black_hole/dal"
	"black_hole/data"
	"black_hole/model"
	"black_hole/strategy"
	"black_hole/trade"
	"fmt"
	"log"
	"time"
)

var (
	zbTradeAPI    *trade.ZbLive
	zbDataAPI     *data.ZbLiveData
	tradeStrategy *strategy.Strategy
)

func Poll() {
	//对于一个策略来：
	//先进入strategy寻找买入或者卖出
	//然后进到funds寻找资金，如果资金不提供资金 策略就不会真正生效
	for {
		config := conf.GetConfig()
		barTime := config.BarTime
		coinName := config.CurCurrency
		//开始进入策略中
		kDataLength := len(zbDataAPI.GetKData()[barTime].StructData)
		//只能取前一天的 比如当前天是：2020-02-10 应该取2020-02-09和之前曲线做出的信号
		tradeStrategy.Params.KlineDatas = zbDataAPI.GetKData()[barTime].StructData[:kDataLength-1]
		latestOperation, err := dal.GetLatestOpertion(coinName)
		if err != nil {
			log.Fatalf("[Poll.GetLatestOperation]. err: %v", err)
			time.Sleep(time.Second)
			continue
		}
		buyOrSell := tradeStrategy.Tick(latestOperation)
		//这个singalKey用时间str作为一个信号，后面轮询到的 都不会去做处理
		singalKey := fmt.Sprintf(constdef.SingalKey, zbDataAPI.GetKData()[barTime].StructData[kDataLength-2].TimeStampStr, config.CurCurrency)
		log.Printf("[Poll]: singalKey: %v, coinName: %v, signal: %v", singalKey, coinName, buyOrSell)
		// openOrder之前重新取配置里的杠杆率，便于动态设置
		zbTradeAPI = trade.SetLever(config, zbTradeAPI)
		trade.OpenOrder(zbTradeAPI, buyOrSell, singalKey, coinName)
		time.Sleep(time.Second * 10)
	}
}

func Alarm(zbDataAPI *data.ZbLiveData) {
	// 按照配置，报当前持仓
	config := conf.GetConfig()
	Interval := config.Interval
	for {
		now := time.Now()
		next := now.Add(time.Minute * 1)
		if next.Minute()%int(Interval) == 0 {
			next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), 0, 0, next.Location())
			t := time.NewTimer(next.Sub(now))
			<-t.C
			trade.BroadAsset(zbTradeAPI)
		}
	}
}

func main() {
	fmt.Println("weclome the black hole fund. the key of wealth")
	fmt.Println("start up data collection system.... please wait!")
	zbDataAPI = data.InitZbLiveData()
	fmt.Println("finish  data collection system")
	fmt.Println("start up trade api system ...")
	zbTradeAPI = trade.NewTradeAPI()
	zbTradeAPI.RealTimeData = zbDataAPI
	fmt.Println("finish trade api system ...")
	fmt.Println("get account info start ...")
	err := zbTradeAPI.GetAccountInfo()
	if err != nil {
		log.Fatalf("[GetAccountInfo]. err: %v", err)
		panic(err.Error())
	}
	fmt.Println("finish account info. enjoy! ...")
	//策略启动 策略参数全部固定 禁止改动 当资金超过200万以上 在考虑改动
	tradeStrategy = strategy.NewStrategy(model.StrategyParms{
		SlowMaBars: 20,
		FastMaBars: 5,
	}, strategy.DoubleMaBuy, strategy.DoubleMaSell)
	//启动后台进程
	fmt.Println("start poll")
	go func() {
		Alarm(zbDataAPI)
	}()
	Poll()
}
