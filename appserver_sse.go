package taproot

import "github.com/highgrav/taproot/sse"

/*
Adds a new Server-Sent Events Hub that the application can write to. Note that, unlike WebSocket Hubs, SSE Hubs are write-only.
The "name" is usually keyed to the user's ID; if you need to discriminate more carefully, then use the user ID plus a meaningful ID.
For example, if you are writing a chat app and you only want the user to get updates for the open chat, you could use "user_id::chat_id"
as the key.
*/
func (srv *AppServer) AddSSEHub(name string) {
	if srv.SSEHubs == nil {
		srv.SSEHubs = make(map[string]*sse.SSEHub)
	}
	if _, ok := srv.SSEHubs[name]; ok {
		return
	}
	b := sse.New(name)
	srv.SSEHubs[name] = b
}
