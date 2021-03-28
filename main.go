package main

import (
	"flag"
	"gateway/dao"
	"gateway/http_proxy_router"
	"gateway/lib"
	"gateway/router"
	"github.com/didi/gendry/scanner"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	config = flag.String("c", "./conf/dev/", "配置文件地址")
	mode   = flag.String("m", "dashboard", "运行模式")
)

func main() {
	flag.Parse()
	lib.InitModule(*config, []string{"base", "mysql", "redis"})
	defer lib.Destroy()
	scanner.SetTagName("json")
	if *mode == "dashboard" {
		router.HttpServerRun()
		// 临时设置gendry解析的tag 在这里
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		router.HttpServerStop()
	}
	if *mode == "proxy" {
		dao.SvcManager.LoadOnceService()
		go func() {
			http_proxy_router.HttpServerRun()
		}()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		router.HttpServerStop()
	}

	log.Fatal("请确认运行模式")


}
