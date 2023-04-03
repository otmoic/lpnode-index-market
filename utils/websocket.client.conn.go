package utils

import (
	"context"
	"log"
	"lp_market/logger"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jakehl/goid"
	"github.com/mcuadros/go-defaults"
)

type WebSocketClientConn struct {
	Id                 string
	ConnectStatus      bool
	Url                string                         // 需要链接的url
	Timeout            int                            // 链接超时时间
	ReconnTimeInterval int                            // 重连的间隔时间 sec
	ReconnMaxTime      int                            // 最大的重连间隔时间 sec
	readChan           chan struct{}                  // 读通道
	writeMessageMutex  sync.Mutex                     // 写入消息的锁
	AutoConn           bool                           // 初始化之后是否自动开始链接
	NoDataDisconnect   bool                           // 无数据时断开链接
	NoDataTimeout      int64                          // 多长时间没有数据之后断开链接
	ws                 *websocket.Conn                // web socket的原始链接
	OnDisconnect       func(code int, message string) // 断开链接的回调函数
	OnConnect          func(conn *WebSocketClientConn)
	OnReconnect        func(connectCount int64, lastError string) // 重新链接后的回调函数
	OnMessage          func(message string)                       // 有消息时候的回调函数
	OnPing             func(conn *WebSocketClientConn)            // 链接正常的定时回调
	PingInterval       int64                                      // 多长时间报告一次 链接状态
	SendPingInterval   int64                                      //多长时间主动发一个Ping的包
	LastSendPingTime   int64                                      //最后一次发送Ping的时间
	ConnectNumber      int64                                      // 重连的次数
	ctx                context.Context
	lastReadTimestamp  int64 // 最后的数据更新时间
	LastPintTime       int64
	HttpHeader         http.Header
	LastError          string //最后一次链接产生的错误
}
type WebSocketClientConnOption struct {
	Url                string // 需要链接的url
	Timeout            int    `default:"30"` // 链接超时时间
	ReconnTimeInterval int    `default:"5"`  // 重连的间隔时间 sec
	ReconnMaxTime      int    `default:"30"` // 最大的重连间隔时间 sec

	AutoConn         bool        `default:"true"` // 初始化之后是否自动开始链接
	NoDataDisconnect bool        `default:"true"` // 无数据时断开链接
	NoDataTimeout    int64       `default:"10"`   // 多长时间没有数据之后断开链接
	SendPingInterval int64       `default:"0"`
	PingInterval     int64       `default:"15"`
	HttpHeader       http.Header // http  header
}

func NewWebSocketClientConn(ctx context.Context, options WebSocketClientConnOption) *WebSocketClientConn {
	if options.Url == "" {
		panic("Url 不能为空")
	}
	config := new(WebSocketClientConnOption)
	defaults.SetDefaults(config)
	if options.Url == "" {
		panic("Url 不能为空")
	}
	config.Url = options.Url
	if options.Timeout > 0 {
		config.Timeout = options.Timeout
	}
	if options.ReconnTimeInterval > 0 {
		config.ReconnTimeInterval = options.ReconnTimeInterval
	}
	if options.PingInterval > 0 {
		config.PingInterval = options.PingInterval
	}
	if options.ReconnMaxTime > 0 {
		config.ReconnMaxTime = options.ReconnMaxTime
	}
	if options.AutoConn != true {
		config.AutoConn = options.AutoConn
	}
	if options.NoDataDisconnect != true {
		config.NoDataDisconnect = options.NoDataDisconnect
	}
	if options.NoDataTimeout > 0 {
		config.NoDataTimeout = options.NoDataTimeout
	}
	if options.HttpHeader != nil {
		config.HttpHeader = options.HttpHeader
	}
	if options.SendPingInterval > 0 {
		config.SendPingInterval = options.SendPingInterval
	}

	conn := &WebSocketClientConn{
		ctx:                ctx,
		Url:                config.Url,
		Timeout:            config.Timeout,
		ReconnTimeInterval: config.ReconnTimeInterval,
		ReconnMaxTime:      config.ReconnMaxTime,
		AutoConn:           config.AutoConn,
		NoDataDisconnect:   config.NoDataDisconnect,
		NoDataTimeout:      config.NoDataTimeout,
		PingInterval:       config.PingInterval,
		SendPingInterval:   config.SendPingInterval,
		HttpHeader:         config.HttpHeader,
		ConnectNumber:      0,
	}
	return conn
}
func (webSocketClientConn *WebSocketClientConn) Initialize() {
	go webSocketClientConn.Run()
}
func (webSocketClientConn *WebSocketClientConn) Run() {
	webSocketClientConn.ConnectStatus = false
	webSocketClientConn.readChan = make(chan struct{})
	_, err := webSocketClientConn.Link() // 链接，并且创建 readchan
	if err != nil {
		logger.Wss.Debug("链接发生了错误", webSocketClientConn.ReconnTimeInterval)
		log.Println("链接发生了错误...,重新链接....", err, webSocketClientConn.Url)
		time.Sleep(time.Second * time.Duration(webSocketClientConn.ReconnTimeInterval))
		webSocketClientConn.Initialize()
		return // 退出协程
	}
	webSocketClientConn.Read(func(message string) {
		if webSocketClientConn.OnMessage != nil {
			webSocketClientConn.OnMessage(message)
		}
	})
	log.Println("链接已经建立完成", webSocketClientConn.Url)
	webSocketClientConn.LastSendPingTime = time.Now().UnixNano() / 1e6
	select {
	case <-webSocketClientConn.readChan:
		log.Println("链接异常中断...", webSocketClientConn.Url)
		webSocketClientConn.ws.Close()
		if webSocketClientConn.OnDisconnect != nil {
			webSocketClientConn.OnDisconnect(1006, "链接断开...")
		}
		if err != nil { // 如果这里出现了Close 错误，有可能导致重连不再继续
			log.Println("这里退出了重连的机制，没有再次 [Initialize]", err)
			return // 退出协程
		}
		time.Sleep(time.Second * time.Duration(webSocketClientConn.ReconnTimeInterval))
		log.Println("重新建立网络链接..", webSocketClientConn.Url)
		webSocketClientConn.Initialize()
		return // 退出协程
	case <-webSocketClientConn.ctx.Done():
		log.Println("链接需要断开.....控制信号要求关闭", webSocketClientConn.Url)
		err := webSocketClientConn.ws.Close()
		if webSocketClientConn.OnDisconnect != nil {
			webSocketClientConn.OnDisconnect(1006, "链接断开...")
		}
		if err != nil {
			return
		}
		return
	}
}
func (webSocketClientConn *WebSocketClientConn) Read(runFun func(message string)) {
	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context, cancelFunc context.CancelFunc) { // 根据时间处理心跳包
		timeInit := time.Now().UnixNano() / 1e6
		webSocketClientConn.LastPintTime = timeInit // 初始化时，先把pingtime设置到现在
		webSocketClientConn.LastSendPingTime = timeInit

		for {
			time.Sleep(time.Millisecond * 500)
			select {
			case <-ctx.Done():
				log.Println("心跳协程....退出.Ctl message", webSocketClientConn.Url)
				return
			default:
				timeNow := time.Now().UnixNano() / 1e6
				if webSocketClientConn.SendPingInterval > 0 && (timeNow-webSocketClientConn.LastSendPingTime) > webSocketClientConn.SendPingInterval*1000 {
					log.Println("send Ping to ", webSocketClientConn.Url)
					webSocketClientConn.SendPing()
					webSocketClientConn.LastSendPingTime = timeNow
				}
				if timeNow-webSocketClientConn.LastPintTime > 1000*webSocketClientConn.PingInterval && webSocketClientConn.ConnectStatus == true {
					if webSocketClientConn.OnPing != nil {
						webSocketClientConn.OnPing(webSocketClientConn)
					}
					webSocketClientConn.LastPintTime = timeNow
				}
				break
			}
		}

	}(ctx, cancel)
	go func(ctx context.Context, cancelFunc context.CancelFunc) { // 读取数据的协程

		v4UUID := goid.NewV4UUID() // 每次read 都产生一个新的uuid networkid
		webSocketClientConn.Id = v4UUID.String()
		webSocketClientConn.ConnectStatus = true
		webSocketClientConn.ConnectNumber++
		if webSocketClientConn.ConnectNumber > 1 && webSocketClientConn.OnReconnect != nil {
			webSocketClientConn.OnReconnect(webSocketClientConn.ConnectNumber, webSocketClientConn.LastError)
		}
		if webSocketClientConn.OnConnect != nil { // 真正开始read ，才算是连接上
			webSocketClientConn.OnConnect(webSocketClientConn)
		}
		for {
			if webSocketClientConn.NoDataDisconnect == true {
				readDeadLineErr := webSocketClientConn.ws.SetReadDeadline(time.Now().Add(time.Second * time.Duration(webSocketClientConn.NoDataTimeout)))
				if readDeadLineErr != nil {
					log.Println("设置deadline发生了错误", readDeadLineErr)
					return
				}
				//webSocketClientConn.ws.SetReadDeadline(time.Now().Add(time.Second * 6))
			}
			_, message, err := webSocketClientConn.ws.ReadMessage()
			timeNow := time.Now().UnixNano() / 1e6
			webSocketClientConn.lastReadTimestamp = timeNow // 读取到消息后，把最后的更新时间更新
			if err != nil {
				webSocketClientConn.ConnectStatus = false
				log.Println("read:", err)
				webSocketClientConn.LastError = err.Error()
				cancel() //控制检测超时的协程关闭 - 控制了 ping
				time.Sleep(time.Millisecond * 1)
				webSocketClientConn.readChan <- struct{}{} //重连的消息
				return
			}
			runFun(string(message))
		}
	}(ctx, cancel)
}

func (webSocketClientConn *WebSocketClientConn) Link() (*websocket.Conn, error) {
	log.Println("connect ", webSocketClientConn.Url)
	ws := websocket.Dialer{}
	// ws.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	logger.Wss.Debug("开始链接...", webSocketClientConn.Url)
	conn, _, err := ws.Dial(webSocketClientConn.Url, webSocketClientConn.HttpHeader)
	if err == nil {
		webSocketClientConn.ws = conn
		//log.Println("链接已经建立成功", webSocketClientConn.Url)
	}

	return conn, err
}
func (webSocketClientConn *WebSocketClientConn) SendTextMessage(message string) {
	if webSocketClientConn.ws != nil && webSocketClientConn.ConnectStatus == true {
		webSocketClientConn.writeMessageMutex.Lock()
		defer webSocketClientConn.writeMessageMutex.Unlock()
		err := webSocketClientConn.ws.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("写入消息发生了错误", err)
			webSocketClientConn.ws.Close() //直接关闭链接
		}
	} else {
		log.Println("写入一个不存在的ws 链接，或者ws链接不可用")
	}

}
func (webSocketClientConn *WebSocketClientConn) SendPing() {
	if webSocketClientConn.ws != nil && webSocketClientConn.ConnectStatus == true {
		webSocketClientConn.writeMessageMutex.Lock()
		defer webSocketClientConn.writeMessageMutex.Unlock()
		err := webSocketClientConn.ws.WriteMessage(websocket.PingMessage, []byte("keepalive"))
		if err != nil {
			log.Println("发送ping消息发生了一个错误", err)
			webSocketClientConn.ws.Close() //直接关闭链接
		}
	} else {
		log.Println("发送ping消息发生了一个错误,写入一个不存在的ws 链接，或者ws链接不可用")
	}

}
