# WEBSOCKETS

Unlike SSEs, websockets are more complex to work with and require additional state. Thus, unlike SSEs, which function 
as a kind of broadcast mechanism, websockets have a relatively complex chain of responsibility.

The `AppServer.HandleWS(brokerName string, createHandler GenerateWSHandler)` function is responsible for managing 
websocket connections. When a connection comes in, it calls `createHandler()` to create a new `websock.IWebSocketHandler`-conformant object, which needs to expose two `chan websock.WSFrame`s (for reading incoming data from the client and 
writing outgoing data) and a `Cancel()` function that should clean up the `websock.IWebSocketHandler` object and remove 
it from memory. `HandleWS(...)` will monitor the websocket, sending data to the incoming channel and, upon receiving 
data on the outgoing channel, writing it back to the websocket. If the websocket is disconnected, it will clean up the 
internal state (the `websock.WSConn` object managing the connection) and call `Cancel()` on the  
`websock.IWebSocketHandler` object.

A simple echo websocket handler is:
~~~
type WebsocketEchoHandler struct {
	SessionID  string
	User       authn.User
	WSIncoming chan websock.WSFrame
	WSOutgoing chan websock.WSFrame
	Done       chan bool
}

func NewWebsocketEchoHandler() websock.IWebSocketHandler {
	return WebsocketEchoHandler{
		SessionID:  "",
		User:       authn.User{},
		WSIncoming: make(chan websock.WSFrame),
		WSOutgoing: make(chan websock.WSFrame),
		Done:       make(chan bool),
	}
}

func (h WebsocketEchoHandler) Cancel() error {
	h.Done <- true
	return nil
}

func (h WebsocketEchoHandler) GetChannels() (wsReader, wsWriter chan websock.WSFrame, err error) {
	return h.WSIncoming, h.WSOutgoing, nil
}

func (h WebsocketEchoHandler) Init(w http.ResponseWriter, r *http.Request) (wsReader, wsWriter chan websock.WSFrame, err error) {
    // Add a header to the websocket response
    w.Header().Add("Content-Type", "text/vnd.turbo-stream.html; charset=utf-8")

	h.SessionID = r.Context().Value(constants.HTTP_CONTEXT_SESSION_KEY).(string)
	// If we had an issue here, we could send back an error on the ResponseWriter and return an error here
	
	h.User, _ = authn.GetUserFromRequest(r)
	tick := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case _ = <-h.Done:
				return
			case C := <-tick.C:
				h.WSOutgoing <- websock.WSFrame{
					Op:   ws.OpText,
					Data: []byte(C.String()),
				}
			case inc := <-h.WSIncoming:
				h.WSOutgoing <- inc
			}
		}
	}()
	return h.WSIncoming, h.WSOutgoing, nil
}
~~~

To implement this, add an endpoint to the AppServer as follows:
~~~
server.AddWSHub("test")
// ... 
server.Handler(http.MethodGet, "/ws", server.HandleWS("test", handlers.NewWebsocketEchoHandler))
~~~