package model

type BackTestDataElement struct {
	Date        string
	OpenPrice   float64
	ClosePrice  float64
	HightPrice  float64
	LowPrice    float64
	TotalProfit float64   //当前总收益
}

type BiasDataElement struct{
	Data string
	Bias float64
}
