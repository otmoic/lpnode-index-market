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
	logger.Httpd.Debug("开始启动Http服务")
	ctrl := &Ctrl{}
	ctrl.RegRouter(httpdRouter)
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		log.Println("无法获得端口的基本配置")
		os.Exit(0)
	}

	panic(fasthttp.ListenAndServe(fmt.Sprintf(":%s", port), httpdRouter.HandleRequest))
}

var httpdInstance *Httpd

// 单例返回  现货市场实例
func GetHttpdInstance() *Httpd {
	once.Do(func() {
		httpdInstance = &Httpd{}
		httpdRouter = routing.New()
	})
	return httpdInstance
}
