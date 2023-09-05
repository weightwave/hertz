package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

var cli *client.Client
var err error

func init() {
	cli, err = client.NewClient(
		client.WithDialTimeout(1*time.Second), // 连接建立超时时间，默认 1s
		//client.WithDialer(standard.NewDialer()),
		client.WithKeepAlive(true),
		client.WithMaxIdleConnDuration(5*time.Second), // 设置空闲连接超时时间，当超时后会关闭该连接，默认10s
		client.WithMaxConnDuration(10*time.Second),    // 设置连接存活的最大时长，超过这个时间的连接在完成当前请求后会被关闭，默认无限长
		client.WithMaxConnWaitTimeout(5*time.Second),  // 设置等待空闲连接的最大时间，默认不等待
		//client.WithTLSConfig(&tls.Config{
		//	InsecureSkipVerify: true,
		//}),
	)
	if err != nil {
		return
	}
}

func main() {
	h := server.Default()
	h.POST("/post", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(200, utils.H{"hel": "wor"})
	})

	h.GET("/post", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(200, utils.H{"hel": "wor"})
	})

	go func() {
		time.Sleep(1 * time.Second)
		for i := 0; i < 1000; i++ {
			go func() {
				req := protocol.AcquireRequest()
				resp := protocol.AcquireResponse()
				err := Post("http://localhost:8888/post", nil, req, resp)
				if err != nil {
					logrus.Error(err.Error())
				}
			}()
		}
	}()

	h.Spin()
}

func Post(url string, header map[string]string, req, resp interface{}) error {

	R := &protocol.Request{}
	RP := &protocol.Response{}

	R.SetMethod(consts.MethodGet)
	R.SetRequestURI(url)
	R.SetHeader("Content-Type", "application/json; charset=UTF-8")
	R.SetHeaders(header)

	defer func() {
		logrus.WithFields(logrus.Fields{
			"R.Header. Header":  string(R.Header.Header()),
			"R.Body":            string(R.Body()),
			"RP.Header. Header": string(RP.Header.Header()),
			"PR.Body":           string(RP.Body()),
		}).Info("请求上游")
	}()

	reqJson, err := json.Marshal(req)
	if err != nil {
		return err
	}
	R.AppendBody(reqJson)

	err = cli.DoTimeout(context.Background(), R, RP, 2*time.Second)
	if err != nil {
		return err
	}

	switch RP.StatusCode() {
	case 200:
		goto end
	case 400:
		return errors.New("400 interfaceError")
	case 401:
		return errors.New("401 Unauthorized")
	case 403:
		return errors.New("403 permissionDenied")
	case 404:
		return errors.New("404 page not found")
	case 429:
		return errors.New("429 trigger current limiting")
	case 503:
		return errors.New("503 Service unavailable")
	default:
		return errors.New("service is gone")
	}

end:

	//fmt.Printf("----->%v %v %v", string(RP.Body()), json.Valid([]byte(RP.Body())), RP.StatusCode())

	if !json.Valid(RP.Body()) {
		return fmt.Errorf("%v", string(RP.Body()))
	}

	err = json.Unmarshal(RP.Body(), &resp)
	if err != nil {
		return err
	}

	return nil

}
