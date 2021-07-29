package backtest

import (
	"black_hole/data"
	"black_hole/model"
	"black_hole/strategy"
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCalBias(t *testing.T,){
	zbLiveData := data.InitZbLiveData()
	kLine := zbLiveData.GetKData()
	structDatas := kLine["1day"].StructData
	DrawBias(structDatas)
}

//回测函数
func TestBackTest(t *testing.T) {
	zbLiveData := data.InitZbLiveData()
	kLine := zbLiveData.GetKData()
	structDatas := kLine["1day"].StructData
	tradeStrategy := strategy.NewStrategy(model.StrategyParms{
		SlowMaBars: 10,
		FastMaBars: 5,
		KlineDatas: structDatas,
	}, strategy.DoubleMaBuy, strategy.DoubleMaSell)
	BackTest(tradeStrategy, 1.5, 1.5, 0.002, 0.0005)
}

//回测函数
func TestBackTestOffline(t *testing.T) {
	structDatas := ReadCsv_ConfigFile_Fun("../btc.csv")
	tradeStrategy := strategy.NewStrategy(model.StrategyParms{
		SlowMaBars: 20,
		FastMaBars: 5,
		KlineDatas: structDatas,
	}, strategy.DoubleMaBuy, strategy.DoubleMaSell)
	BackTest(tradeStrategy, 1.0, 1.0, 0.002, 0.0005)
}

// 游戏读取数据，读取游戏配置数据
func ReadCsv_ConfigFile_Fun(fileName string) []model.KLineBar {
	csvFile, _ := os.Open(fileName)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	reader.LazyQuotes = true
	index := 0
	var structKlines []model.KLineBar
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			fmt.Println("err: ", error)
			time.Sleep(time.Second)
			panic("read failed")
		}
		index++
		if index == 1 {
			continue
		}
		yearIndex := strings.Index(line[0], "年")
		year, _ := strconv.ParseInt(line[0][:yearIndex], 10, 64)
		if year < 2013 {
			continue
		}
		if year > 2021 {
			continue
		}
		line[2] = strings.Replace(line[2], ",", "", -1)
		line[1] = strings.Replace(line[1], ",", "", -1)
		line[3] = strings.Replace(line[3], ",", "", -1)
		line[4] = strings.Replace(line[4], ",", "", -1)
		openPrice, err1 := strconv.ParseFloat(line[2], 64)
		closePrice, err2 := strconv.ParseFloat(line[1], 64)
		if err1 != nil || err2 != nil {
			log.Fatalf("err1: %v, err2: %v, openPrice: %v, closePrice: %v", err1, err2, line[2], line[1])
		}
		hightPrice, err3 := strconv.ParseFloat(line[3], 64)
		lowPrice, err4 := strconv.ParseFloat(line[4], 64)
		if err3 != nil || err4 != nil {
			log.Fatalf("err1: %v, err2: %v, openPrice: %v, closePrice: %v", err3, err4, line[3], line[4])
		}
		barData := model.KLineBar{
			OpenPrice:    openPrice,
			ClosePrice:   closePrice,
			LowPrice:     lowPrice,
			HighPrice:    hightPrice,
			Volum:        0,
			TimeStamp:    0,
			TimeStampStr: line[0],
			CoinName:     "btc",
		}
		structKlines = append(structKlines, barData)
	}
	for i, j := 0, len(structKlines)-1; i < j; i, j = i+1, j-1 {
		structKlines[i], structKlines[j] = structKlines[j], structKlines[i]
	}
	return structKlines
}

func TestReadCsv(t *testing.T) {
	structData := ReadCsv_ConfigFile_Fun("../btc.csv")
	log.Printf("structData: %v", structData)
	time.Sleep(time.Second)
}
