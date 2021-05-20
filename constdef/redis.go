package constdef

import "strings"

const (
	SingalKey             = "singal_%v_%v"             //{timestamp_type}：时间戳 | 杠杆品种（高、低） 对应的状态
	LatestSingalStatusKey = "latest_singal_status_key" //最近信号状态
	LatestSingalKey       = "latest_singal_key"        //最新信号key
	OrderDetailKey        = "order_detail_key_%v"      //每个订单详细内容
	LoanInfo              = "loan_coin_name"           // 借款情况
	LatestOperationKey    = "latest_operation_key"     //最近操作的key
	LatestKey             = "latest_key_%v"            // coin_name
)

//是否含有详细订单信息
func ContainOrderDetailKey(str string) bool {
	if strings.Contains(str, "order_detail_key") {
		return true
	}
	return false
}
func GetOrderIdByKey(str string) string {
	return strings.Trim(str, "order_detail_key")
}