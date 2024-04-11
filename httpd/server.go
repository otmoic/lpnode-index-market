package httpd

import (
	"fmt"
	"log"
	"lp_market/logger"
	"os"
	"sync"

	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

var httpdRouter *routing.Router

var once sync.Once

type Httpd struct {
}

func (httpd *Httpd) Init() {
	logger.Httpd.Debug("start http service")
	ctrl := &Ctrl{}
	ctrl.RegRouter(httpdRouter)
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		log.Println("unable to obtain basic configuration for port")
		os.Exit(0)
	}

	panic(fasthttp.ListenAndServe(fmt.Sprintf(":%s", port), httpdRouter.HandleRequest))
}

var httpdInstance *Httpd

func GetHttpdInstance() *Httpd {
	once.Do(func() {
		httpdInstance = &Httpd{}
		httpdRouter = routing.New()
	})
	return httpdInstance
}
