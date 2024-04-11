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
	Url                string                         // URL required for connection
	Timeout            int                            // Connection timeout
	ReconnTimeInterval int                            // Reconnection interval in seconds
	ReconnMaxTime      int                            // Maximum reconnection interval in seconds
	readChan           chan struct{}                  // Read channel
	writeMessageMutex  sync.Mutex                     // Lock for writing messages
	AutoConn           bool                           // Whether to automatically start connecting after initialization
	NoDataDisconnect   bool                           // Disconnect when no data is received
	NoDataTimeout      int64                          // Time in seconds after which to disconnect if no data is received
	ws                 *websocket.Conn                // Raw WebSocket connection
	OnDisconnect       func(code int, message string) // Callback function for disconnections
	OnConnect          func(conn *WebSocketClientConn)
	OnReconnect        func(connectCount int64, lastError string) // Callback function after reconnection
	OnMessage          func(message string)                       // Callback function when a message is received
	OnPing             func(conn *WebSocketClientConn)            // Periodic callback indicating a healthy connection
	PingInterval       int64                                      // Interval in seconds at which to report connection status
	SendPingInterval   int64                                      // Interval in seconds at which to send an active Ping packet
	LastSendPingTime   int64                                      // Unix timestamp of the last sent Ping
	ConnectNumber      int64                                      // Number of reconnect attempts
	ctx                context.Context
	lastReadTimestamp  int64 // Unix timestamp of the last data update
	LastPintTime       int64
	HttpHeader         http.Header
	LastError          string // Error from the last connection attempt
}
type WebSocketClientConnOption struct {
	Url                string // URL required for connection
	Timeout            int    `default:"30"` // Connection timeout
	ReconnTimeInterval int    `default:"5"`  // Reconnection interval in seconds
	ReconnMaxTime      int    `default:"30"` // Maximum reconnection interval in seconds

	AutoConn         bool        `default:"true"` // Whether to automatically start connecting after initialization
	NoDataDisconnect bool        `default:"true"` // Disconnect when no data is received
	NoDataTimeout    int64       `default:"10"`   // Time in seconds after which to disconnect if no data is received
	SendPingInterval int64       `default:"0"`
	PingInterval     int64       `default:"15"`
	HttpHeader       http.Header // http  header
}

func NewWebSocketClientConn(ctx context.Context, options WebSocketClientConnOption) *WebSocketClientConn {
	if options.Url == "" {
		panic("Url cannot be empty")
	}
	config := new(WebSocketClientConnOption)
	defaults.SetDefaults(config)
	if options.Url == "" {
		panic("Url cannot be empty")
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
	if !options.AutoConn {
		config.AutoConn = options.AutoConn
	}
	if !options.NoDataDisconnect {
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
	_, err := webSocketClientConn.Link() // Connect and create readchan
	if err != nil {
		logger.Wss.Debug("Connection error occurred", webSocketClientConn.ReconnTimeInterval)
		log.Println("Connection error occurred..., attempting to reconnect...", err, webSocketClientConn.Url)
		time.Sleep(time.Second * time.Duration(webSocketClientConn.ReconnTimeInterval))
		webSocketClientConn.Initialize()
		return //  Exit goroutine
	}
	webSocketClientConn.Read(func(message string) {
		if webSocketClientConn.OnMessage != nil {
			webSocketClientConn.OnMessage(message)
		}
	})
	log.Println("Connection established successfully", webSocketClientConn.Url)
	webSocketClientConn.LastSendPingTime = time.Now().UnixNano() / 1e6
	select {
	case <-webSocketClientConn.readChan:
		log.Println("Connection interrupted unexpectedly...", webSocketClientConn.Url)
		webSocketClientConn.ws.Close()
		if webSocketClientConn.OnDisconnect != nil {
			webSocketClientConn.OnDisconnect(1006, "Connection disconnected...")
		}
		if err != nil { // If there's an error here on Close, it could prevent further reconnections
			log.Println("Exiting reconnection mechanism due to Close error. Not initializing again.", err)
			return // Exit goroutine
		}
		time.Sleep(time.Second * time.Duration(webSocketClientConn.ReconnTimeInterval))
		log.Println("Attempting to reestablish network connection...", webSocketClientConn.Url)
		webSocketClientConn.Initialize()
		return //  Exit goroutine
	case <-webSocketClientConn.ctx.Done():
		log.Println("Connection needs to be terminated... Control signal requests closure", webSocketClientConn.Url)
		err := webSocketClientConn.ws.Close()
		if webSocketClientConn.OnDisconnect != nil {
			webSocketClientConn.OnDisconnect(1006, "Connection disconnected...")
		}
		if err != nil {
			return
		}
		return
	}
}
func (webSocketClientConn *WebSocketClientConn) Read(runFun func(message string)) {
	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context, cancelFunc context.CancelFunc) { // Handle heartbeats based on time
		timeInit := time.Now().UnixNano() / 1e6
		webSocketClientConn.LastPintTime = timeInit // Initialize ping time to now
		webSocketClientConn.LastSendPingTime = timeInit

		for {
			time.Sleep(time.Millisecond * 500)
			select {
			case <-ctx.Done():
				log.Println("Heartbeat goroutine exiting... Ctrl message", webSocketClientConn.Url)
				return
			default:
				timeNow := time.Now().UnixNano() / 1e6
				if webSocketClientConn.SendPingInterval > 0 && (timeNow-webSocketClientConn.LastSendPingTime) > webSocketClientConn.SendPingInterval*1000 {
					log.Println("send Ping to ", webSocketClientConn.Url)
					webSocketClientConn.SendPing()
					webSocketClientConn.LastSendPingTime = timeNow
				}
				if timeNow-webSocketClientConn.LastPintTime > 1000*webSocketClientConn.PingInterval && webSocketClientConn.ConnectStatus {
					if webSocketClientConn.OnPing != nil {
						webSocketClientConn.OnPing(webSocketClientConn)
					}
					webSocketClientConn.LastPintTime = timeNow
				}
				break
			}
		}

	}(ctx, cancel)
	go func(ctx context.Context, cancelFunc context.CancelFunc) { // Goroutine to read data

		v4UUID := goid.NewV4UUID() // Generate a new UUID network ID for each read
		webSocketClientConn.Id = v4UUID.String()
		webSocketClientConn.ConnectStatus = true
		webSocketClientConn.ConnectNumber++
		if webSocketClientConn.ConnectNumber > 1 && webSocketClientConn.OnReconnect != nil {
			webSocketClientConn.OnReconnect(webSocketClientConn.ConnectNumber, webSocketClientConn.LastError)
		}
		if webSocketClientConn.OnConnect != nil { // Consider connected once reading actually begins
			webSocketClientConn.OnConnect(webSocketClientConn)
		}
		for {
			if webSocketClientConn.NoDataDisconnect {
				readDeadLineErr := webSocketClientConn.ws.SetReadDeadline(time.Now().Add(time.Second * time.Duration(webSocketClientConn.NoDataTimeout)))
				if readDeadLineErr != nil {
					log.Println("An error occurred while setting the deadline:", readDeadLineErr)
					return
				}
				//webSocketClientConn.ws.SetReadDeadline(time.Now().Add(time.Second * 6))
			}
			_, message, err := webSocketClientConn.ws.ReadMessage()
			timeNow := time.Now().UnixNano() / 1e6
			webSocketClientConn.lastReadTimestamp = timeNow // Update the last update time after receiving a message
			if err != nil {
				webSocketClientConn.ConnectStatus = false
				log.Println("read:", err)
				webSocketClientConn.LastError = err.Error()
				cancel() // Close the timeout detection goroutine - controls ping
				time.Sleep(time.Millisecond * 1)
				webSocketClientConn.readChan <- struct{}{} // Reconnect message
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
	logger.Wss.Debug("Starting connection...", webSocketClientConn.Url)
	conn, _, err := ws.Dial(webSocketClientConn.Url, webSocketClientConn.HttpHeader)
	if err == nil {
		webSocketClientConn.ws = conn
		//log.Println("Connection established successfully", webSocketClientConn.Url)
	}

	return conn, err
}
func (webSocketClientConn *WebSocketClientConn) SendTextMessage(message string) {
	if webSocketClientConn.ws != nil && webSocketClientConn.ConnectStatus {
		webSocketClientConn.writeMessageMutex.Lock()
		defer webSocketClientConn.writeMessageMutex.Unlock()
		err := webSocketClientConn.ws.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("An error occurred while sending a message:", err)
			webSocketClientConn.ws.Close() // Close the connection directly
		}
	} else {
		log.Println("Attempting to write to a non-existent or invalid ws connection")
	}

}
func (webSocketClientConn *WebSocketClientConn) SendPing() {
	if webSocketClientConn.ws != nil && webSocketClientConn.ConnectStatus {
		webSocketClientConn.writeMessageMutex.Lock()
		defer webSocketClientConn.writeMessageMutex.Unlock()
		err := webSocketClientConn.ws.WriteMessage(websocket.PingMessage, []byte("keepalive"))
		if err != nil {
			log.Println("An error occurred while sending a ping message:", err)
			webSocketClientConn.ws.Close() // close ws
		}
	} else {
		log.Println("An error occurred while sending a ping message, attempting to write to a non-existent or invalid ws connection")
	}

}
