package model

//聚合体 所有的参数

type StrategyParms struct {
	SlowMaBars int
	FastMaBars int
	KlineDatas []KLineBar

	Name       string //策略名字
	IsBackTest bool   //是否是回测
}
