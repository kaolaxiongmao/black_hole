package data

//取数据相关接口
type Data interface {
	GetDepthData() (interface{}, error) //取数据相关接口
	GetKData() (interface{}, error)     //取数据
}
