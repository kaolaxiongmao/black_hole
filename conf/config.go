package conf

import (
	"bufio"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

//这里字段必须是大写开头,因为接下来需要用反射给属性赋值,大写才可以取到
type Config struct {
	CurCurrency   string
	MultipleLever float64
	EmptyLever    float64
	LowAccessKey  string
	LowSecretKey  string
	LowPwd        string
	BarTime       string
	Interval      int64
}

func GetConfig() Config {
	c := &Config{}
	cr := reflect.ValueOf(c).Elem()
	f, err := os.Open("./config.conf")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	s := bufio.NewScanner(f)
	for s.Scan() {
		var str = s.Text()
		var index = strings.Index(str, "=")
		var key = str[0:index]
		var value = str[index+1:]
		//因为杠杆率是float,所以我们这里要将截取的string强转成float
		if strings.Contains(key, "Lever") {
			var i, err = strconv.ParseFloat(value, 64)
			if err != nil {
				panic(err)
			}
			//通过反射将字段设置进去
			cr.FieldByName(key).Set(reflect.ValueOf(i))
		} else if strings.Contains(key, "Interval") {
			var i, err = strconv.ParseInt(value, 10, 64)
			if err != nil {
				panic(err)
			}
			cr.FieldByName(key).Set(reflect.ValueOf(i))
		} else {
			//通过反射将字段设置进去
			cr.FieldByName(key).Set(reflect.ValueOf(value))
		}

	}
	err = s.Err()
	if err != nil {
		log.Fatal(err)
	}
	//返回Config结构体变量
	return *c
}
