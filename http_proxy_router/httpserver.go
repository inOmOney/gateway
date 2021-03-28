package http_proxy_router

import (
	"gateway/lib"
	"net/http"
)

func HttpServerRun(){
	r := InitRouter()
	httpserver := &http.Server{
		Addr: lib.GetStringConf("proxy.http.addr"),
		Handler: r,
		//todo 还有些连接时间配置
	}
	httpserver.ListenAndServe()
}
