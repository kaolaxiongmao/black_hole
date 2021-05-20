package dal

import (
	"fmt"
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	val, err := OrderStorage.Set("test", -1, 0).Result()
	fmt.Println("val: ", val, " err: ", err)
	time.Sleep(time.Second)
}

func TestHSet(t *testing.T) {
	SetNewOrder("test_111", "122222", 12.22, 11, 1)
	time.Sleep(time.Second)
}

