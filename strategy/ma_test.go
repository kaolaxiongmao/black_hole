package strategy

import (
	"black_hole/data"
	"fmt"
	"github.com/markcheno/go-talib"
	"testing"
)

//每个测试策略需要严格执行2个： 代码正确性/以及历史数据测试 历史回测进入到backtest回测框架来做测试
//测试ma运作的正确性
func TestMa(t *testing.T) {
	zbLiveData := data.InitZbLiveData()
	kLine := zbLiveData.GetKData()
	structDatas := kLine["1day"].StructData
	var closePrices []float64
	for _, close := range structDatas {
		closePrices = append(closePrices, close.ClosePrice)
	}
	startIndex := 0
	for i := 20; i <= len(closePrices); i++ {
		ma20 := talib.Sma(closePrices[startIndex:i], 20)
		fmt.Println("ma20: ", ma20)
		startIndex++
	}
}

func TestMAb(t *testing.T) {
	a := 0.0
	a = 2088 + 2116 + 2109 + 2145 +  2189 + 2320 + 2313 + 2368 + 2456 + 2601
	fmt.Println(a / 10)
	fmt.Println(float64(2601-2270)/2270)
}