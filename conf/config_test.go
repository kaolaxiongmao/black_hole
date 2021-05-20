package conf

import (
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	str, _ := os.Getwd()
	fmt.Println(str)
	c := GetConfig()
	fmt.Println(c)
	fmt.Println(c.Interval)
}

