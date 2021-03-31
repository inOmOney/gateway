package reverse_proxy

import (
	"gateway/reverse_proxy/load_balance"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var transport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second, //连接超时
		KeepAlive: 30 * time.Second, //长连接超时时间
	}).DialContext,
	MaxIdleConns:          100,              //最大空闲连接
	IdleConnTimeout:       90 * time.Second, //空闲超时时间
	TLSHandshakeTimeout:   10 * time.Second, //tls握手超时时间
	ExpectContinueTimeout: 1 * time.Second,  //100-continue 超时时间
}

func GetReverseProxy(c *gin.Context, lb load_balance.LoadBalance) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		target, err := url.Parse(lb.Get(c.Request.RemoteAddr + c.Request.URL.Path))
		if err != nil {
			panic(err)
		}
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		//req.URL.Path = "/base"
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "user-agent")
		}
	}
	return &httputil.ReverseProxy{
		Director:  director,
		Transport: transport,
	}
}
