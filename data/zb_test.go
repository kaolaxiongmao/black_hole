package data

import (
	"testing"
	"time"
)

func TestInitZbLiveData(t *testing.T) {
	InitZbLiveData()
	//开始获取数据
	time.Sleep(time.Second * 100)
}
