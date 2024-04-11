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

func main() {
	log.Printf("init...")
	database_config.Init()
	database.InitConnect("main", nil)

	marketCenter := &market.MarketCenter{}

	dCtx, dCancel := context.WithCancel(context.Background())
	go func() {
		err := state.GetStateDbInstance().Init(dCtx)
		if err != nil {
			logger.MainMessage.Errorf("error starting database:%s", err)
		}
	}()
	// initializing spot market data
	spotCtx, spotCancel := context.WithCancel(context.Background())
	marketCenter.InitSpot(spotCtx, spotCancel)
	//  initializing u-based market data
	uSwapCtx, uSwapCancel := context.WithCancel(context.Background())
	marketCenter.InitUsdtSwap(uSwapCtx, uSwapCancel)
	// initializing co-based market data
	cSwapCtx, cSwapCancel := context.WithCancel(context.Background())
	marketCenter.InitCoinSwap(cSwapCtx, cSwapCancel)
	// funding rate
	uSwapFundingRateCtx, uSwapFundingRateCancel := context.WithCancel(context.Background())
	marketCenter.InitFundingRate(uSwapFundingRateCtx, uSwapFundingRateCancel)

	statusReport := statusreport.NewStatusReport()
	statusReport.UpdateStatus()          // collecting periodically into memory
	go statusReport.IntervalReport()     // periodically storing
	go redisbus.GetRedisBus().SubEvent() // subscribing to system events

	go marketrefresh.RefreshSpot() // processing system events
	go func() {                    // upon initialization, subscribe to all currency pair market data in the database
		time.Sleep(time.Second * 5) // after startup, initially subscribe to all spot market data
		logger.MainMessage.Warn("initialize loading of currency pairs")
		redisbus.GetRedisBus().EventList <- &redisbus.LPSystemNoticeEventItem{Str: `{"type":"systemInit","payload":"{}"}`}
	}()
	// time.Sleep(time.Second * 100 * 100)
	// used to start http server
	go func() {
		httpd.GetHttpdInstance().Init() // invoke httpd and initialize
	}()
	go func() {
		for {
			time.Sleep(time.Second * 20)
			runtime.GC()
		}
	}()

	// begin listening for operating system signals, program does not exit
	sysMonit := make(chan os.Signal, 100)
	signal.Notify(sysMonit, syscall.SIGINT, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	sig := <-sysMonit
	dCancel()     // notify database module
	spotCancel()  // notify spot module
	uSwapCancel() // notify u-based contract module
	cSwapCancel() // notify coin-margined contract module
	uSwapFundingRateCancel()
	logger.MainMessage.Logger.Debug("preparing to exit system...")
	time.Sleep(time.Second * 2) // wait for goroutines to exit
	log.Println("program has exited normally......", sig)
	os.Exit(0)

}
