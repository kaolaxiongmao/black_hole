package strategy

import (
	"black_hole/constdef"
	"black_hole/model"
)

type Strategy struct {
	Buy    func(params *model.StrategyParms) constdef.TradeType
	Sell   func(params *model.StrategyParms) constdef.TradeType
	Params model.StrategyParms
}
//新建一个策略入口
//策略所需参数
//买入触发器
//卖出触发器
func NewStrategy(params model.StrategyParms, buy func(params *model.StrategyParms) constdef.TradeType, sell func(params *model.StrategyParms) constdef.TradeType) *Strategy {
	tradeStrategy := &Strategy{
		Buy:    buy,
		Sell:   sell,
		Params: params,
	}
	return tradeStrategy
}

//带状态Tick
//一旦策略产生了买入动作以后，接下来就是一直在等机会卖出。
func (s *Strategy) Tick(latestSingalFlag constdef.TradeType) constdef.TradeType {
	var re constdef.TradeType
	if latestSingalFlag != constdef.Buy {
		re = s.Buy(&s.Params)
		if re == constdef.Buy {
			return re
		}
	}
	if latestSingalFlag != constdef.Sell {
		//查看是否卖出
		re = s.Sell(&s.Params)
		if re == constdef.Sell {
			return re
		}
	}
	return re
}


