package logger

import (
	"fmt"
	"os"
	"runtime"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"github.com/zput/zxcTool/ztLog/zt_formatter"
)

var Log *logrus.Logger
var MainMessage *logrus.Entry
var SpotMarket *logrus.Entry
var Httpd *logrus.Entry
var StateDb *logrus.Entry
var Wss *logrus.Entry
var Orderbook *logrus.Entry
var USwapMarket *logrus.Entry
var CSwapMarket *logrus.Entry
var FundingRate *logrus.Entry
var Config *logrus.Entry

func init() {
	var formatter = &zt_formatter.ZtFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := f.File
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
		Formatter: nested.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			//HideKeys: true,
			FieldsOrder: []string{"M", "category"},
		},
	}
	l := logrus.New()
	l.SetReportCaller(true)
	l.SetLevel(logrus.DebugLevel)
	if formatter != nil {
		l.SetFormatter(formatter)
	}
	l.WithField("c", "BscLightNode")
	l.SetOutput(os.Stdout)
	Log = l
	SpotMarket = Log.WithField("M", "SpotMarket🐶")
	MainMessage = Log.WithField("M", "Main🌗")
	Httpd = Log.WithField("M", "Httpd🌗")
	StateDb = Log.WithField("M", "StateDb🌗")
	USwapMarket = Log.WithField("M", "USwapMarket")
	Wss = Log.WithField("M", "StateDb🌗")
	CSwapMarket = Log.WithField("M", "CSwapMarket")
	Orderbook = Log.WithField("M", "ORDERBOOK")
	FundingRate = Log.WithField("M", "FundingRate")
	Config = Log.WithField("M", "Config")

}
