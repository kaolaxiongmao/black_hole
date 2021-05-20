package chart

import (
	"black_hole/model"
	"testing"
	"time"
)

func TestEntry(t *testing.T) {
	datas := make([]model.BackTestDataElement, 500)
	for i := 0; i < 500; i++ {
		datas[i] = model.BackTestDataElement{
			Date:        "2020-01-01",
			OpenPrice:   0,
			ClosePrice:  0,
			HightPrice:  0,
			LowPrice:    0,
			TotalProfit: float64(i)*float64(i),
		}
	}
	StartHttpServer("宙斯计划", "btc", datas)
	time.Sleep(time.Second)
}

