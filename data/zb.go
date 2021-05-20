package data

import (
	"black_hole/constdef"
	"black_hole/model"
	"black_hole/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

type ZbLiveData struct {
	depth     model.DepthData
	kLine     map[string]*model.KLineData
	kLineLock sync.Mutex
	depthLock sync.Mutex
}

func (zb *ZbLiveData) GetDepthData() model.DepthData {
	zb.depthLock.Lock()
	defer zb.depthLock.Unlock()
	return zb.depth
}

func (zb *ZbLiveData) SetDepthData(market string, size int) {
	url := constdef.DataZbUrl + "depth?" + fmt.Sprintf("market=%v", market) + fmt.Sprintf("&size=%v", size)
	data, err := utils.HttpGet(url)
	if err != nil {
		utils.SendLarkNoticeToDefaultUser(context.Background(), "❌获取深度信息失败")
		log.Fatalf("[GetDepthData.HttpGet], err: %s", err.Error())
		return
	}
	zb.depthLock.Lock()
	defer zb.depthLock.Unlock()
	jsonErr := json.Unmarshal([]byte(data), &zb.depth)
	if jsonErr != nil {
		log.Fatalf("[GetDepthData.Unmarshal]: %v", jsonErr)
	}
	zb.depth.LocalTimeStamp = time.Now().Unix()
}

func (zb *ZbLiveData) GetKData() map[string]*model.KLineData{
	zb.kLineLock.Lock()
	defer zb.kLineLock.Unlock()
	return zb.kLine
}

func (zb *ZbLiveData) SetKData(market string, kDataType string) {
	url := constdef.DataZbUrl + "kline?" + fmt.Sprintf("market=%v", market) + fmt.Sprintf("&type=%s", kDataType)
	data, err := utils.HttpGet(url)
	if err != nil {
		utils.SendLarkNoticeToDefaultUser(context.Background(), "❌获取k线数据失败")
		log.Fatalf("[GetKData.HttpGet], err: %s", err.Error())
		return
	}
	jsonErr := json.Unmarshal([]byte(data), zb.kLine[kDataType])
	if jsonErr != nil {
		log.Fatalf("[GetKData.Unmarshal]: %v", jsonErr)
		return
	}
	zb.kLineLock.Lock()
	defer zb.kLineLock.Unlock()
	for _, data := range zb.kLine[kDataType].Data {
		bar := model.KLineBar{}
		bar.TimeStamp = int64(data[0]) / 1000
		tm := time.Unix(bar.TimeStamp, 0)
		tm.Format("2006-01-02 15:04:05")
		bar.Date = tm
		bar.TimeStampStr = tm.Format("2006-01-02 15:04:05")
		bar.OpenPrice = data[1]
		bar.HighPrice = data[2]
		bar.LowPrice = data[3]
		bar.ClosePrice = data[4]
		bar.Volum = data[5]
		bar.CoinName = market
		zb.kLine[kDataType].StructData = append(zb.kLine[kDataType].StructData, bar)
	}
}

func InitZbLiveData() *ZbLiveData {
	data := &ZbLiveData{}
	data.kLine = make(map[string]*model.KLineData)
	for _, bar := range constdef.BarLine {
		data.kLine[bar] = &model.KLineData{}
	}
	cronJobZb(data)
	return data
}

func cronJobZb(data *ZbLiveData) {
	depthDataTick := time.NewTicker(constdef.DepthFreq)
	kDataTick := time.NewTicker(constdef.KFreq)
	go func() {
		for {
			select {
			case <-depthDataTick.C:
				market := constdef.GetCurMarket()
				data.SetDepthData(market, 5)
			}
		}
	}()
	//init data过程
	market := constdef.GetCurMarket()
	for _, bar := range constdef.BarLine {
		data.SetKData(market, bar)
		time.Sleep(time.Second * 2)
	}

	go func() {
		for {
			select {
			case <-kDataTick.C:
				market := constdef.GetCurMarket()
				for _, bar := range constdef.BarLine {
					data.SetKData(market, bar)
					time.Sleep(time.Second * 2)
				}
			}
		}
	}()
}

