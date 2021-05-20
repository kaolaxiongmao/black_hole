package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func GetReqTime() int64 {
	return time.Now().UnixNano() / 1e6
}

//sha加密算法
func Sha1Algro(data string) string {
	sha1 := sha1.New()
	sha1.Write([]byte(data))
	return hex.EncodeToString(sha1.Sum([]byte("")))
}

//hmac加密算法
func Hmac(key, data string) string {
	hmac := hmac.New(md5.New, []byte(key))
	hmac.Write([]byte(data))
	return hex.EncodeToString(hmac.Sum([]byte("")))
}

//按照升序排序
func JoinStringsInASCII(data map[string]string) string {
	var keyList []string
	for key, _ := range data {
		keyList = append(keyList, key)
	}
	sort.Strings(keyList)
	for index, key := range keyList {
		keyList[index] = key + "=" + data[key]
	}
	return strings.Join(keyList, "&")
}

func HttpGet(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("[HttpGet.Get]: %v,", err.Error())
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("[HttpGet.ReadAll]: %v,", err.Error())
		return "", err
	}
	return string(body), nil
}

func HttpGetWithTimeout(url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("[HttpGetWithTimeout]: err: %s", err.Error())
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PostmanRuntime/7.6.0")
	req.Header.Set("Accpet", "*/*")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("[HttpGetWithTimeout.Do]: err: %s", err.Error())
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("[HttpGet.ReadAll]: %v,", err.Error())
		return "", err
	}
	return string(body), nil
	// ...
}

//按照最小单位来确定一个float值
func MinUnit(origin float64, minUnit float64) float64 {
	return float64(int(origin/minUnit)) * minUnit
}


func StrToFloat64(str string) float64 {
	val, _ := strconv.ParseFloat(str, 10)
	return val
}

func StrToInt64(str string) int64 {
	val, _ := strconv.ParseInt(str, 10, 64)
	return val
}


