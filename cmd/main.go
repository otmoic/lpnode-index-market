package main

import (
	"context"
	"log"
	"lp_market/database_config"
	"lp_market/httpd"
	"lp_market/logger"
	"lp_market/market"
	marketrefresh "lp_market/market_refresh"
	database "lp_market/mongo_database"
	redisbus "lp_market/redis_bus"
	"lp_market/state"
	statusreport "lp_market/status_report"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// @todo 这里可以根据参数来选择启动的市场再设置币对，规避单机处理性能问题
func main() {
	log.Printf("init...")
	database_config.Init()
	database.InitConnect("main", nil)

	marketCenter := &market.MarketCenter{}
	// 启动 stateDb的处理
	dCtx, dCancel := context.WithCancel(context.Background())
	go func() {
		err := state.GetStateDbInstance().Init(dCtx)
		if err != nil {
			logger.MainMessage.Errorf("启动数据库发生了错误:%s", err)
		}
	}()
	// 初始化现货的行情
	spotCtx, spotCancel := context.WithCancel(context.Background())
	marketCenter.InitSpot(spotCtx, spotCancel)
	// 初始化U本位行情
	uSwapCtx, uSwapCancel := context.WithCancel(context.Background())
	marketCenter.InitUsdtSwap(uSwapCtx, uSwapCancel)
	//初始化币本位行情
	cSwapCtx, cSwapCancel := context.WithCancel(context.Background())
	marketCenter.InitCoinSwap(cSwapCtx, cSwapCancel)
	// 资金费率
	uSwapFundingRateCtx, uSwapFundingRateCancel := context.WithCancel(context.Background())
	marketCenter.InitFundingRate(uSwapFundingRateCtx, uSwapFundingRateCancel)

	statusReport := statusreport.NewStatusReport()
	statusReport.UpdateStatus()          // 定时收集到内存中
	go statusReport.IntervalReport()     // 定时存储
	go redisbus.GetRedisBus().SubEvent() // 订阅系统事件

	go marketrefresh.RefreshSpot() // 处理系统事件
	go func() {                    // 初始化时，订阅数据库中所有的币对行情
		time.Sleep(time.Second * 5) // 启动后，先订阅所有的现货行情
		logger.MainMessage.Warn("初始化加载币对")
		redisbus.GetRedisBus().EventList <- &redisbus.LPSystemNoticeEventItem{Str: `{"type":"systemInit","payload":"{}"}`}
	}()
	// time.Sleep(time.Second * 100 * 100)
	// 用于启动http服务器
	go func() {
		httpd.GetHttpdInstance().Init() // 调用httpd 并初始化
	}()
	go func() {
		for {
			time.Sleep(time.Second * 20)
			runtime.GC()
		}
	}()

	// 开始监听操作系统信号，程序不退出
	sysMonit := make(chan os.Signal, 100)
	signal.Notify(sysMonit, syscall.SIGINT, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	sig := <-sysMonit
	dCancel()     // 通知数据库模块
	spotCancel()  // 通知现货模块
	uSwapCancel() // 通知u本位合约
	cSwapCancel() // 通知币本位模块
	uSwapFundingRateCancel()
	logger.MainMessage.Logger.Debug("准备退出系统....")
	time.Sleep(time.Second * 2) //等协程退出
	log.Println("程序已经正常退出.......", sig)
	os.Exit(0)

}
