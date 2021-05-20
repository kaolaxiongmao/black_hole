package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type TextContent struct {
	Text string `json:"text"`
}

type LarkRequest struct {
	Ctx         context.Context `json:"-"`
	MessageType string          `json:"msg_type" validate:"omitempty"`
	Title       string          `json:"title"`
	LogID       string          `json:"-"`
	Text        string          `json:"text" validate:"omitempty"`
	Content     TextContent     `json:"content" validate:"omitempty"`
	version     int
}

func SendLarkNotice(ctx context.Context, msg string) {
	url := "https://open.feishu.cn/open-apis/bot/v2/hook/a577137f-264e-440b-b1c0-360aa677b818"
	// req.SendLarkWithConfig("default_url")
	req := LarkRequest{
		Ctx:         ctx,
		MessageType: "text",
		Title:       "",
		Content:     TextContent{msg},
		version:     2,
	}
	send(ctx, url, req)

}

func SendLarkNoticeToDefaultUser(ctx context.Context, msg string) {
	tm := time.Now()
	timeStr := tm.Format("2006-01-02 15:04:05")
	msg = fmt.Sprintf("✅当前时间:%v\n", timeStr) + msg
	SendLarkNotice(ctx, msg)
}

func send(ctx context.Context, url string, msg LarkRequest) (*http.Response, error) {
	// make content
	b, _ := json.Marshal(msg)
	req, _ := http.NewRequest("POST", url, strings.NewReader(string(b)))
	log.Printf("to send lark message:%+v", string(b))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("send lark fail,err:%+v", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
		ioutil.ReadAll(resp.Body)
	}
	return resp, err
}
