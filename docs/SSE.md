# Server-Sent Events

Taproot makes it simple to add user-specific server-sent events to a server. The `AppServer.SSEHubs` map allows 
you to add multiple SSE "hubs" to the server. You can send messages (of the type `sse.SSEEvent`) to a hub, to one, many, 
or all connected users (or, if using a custom handler, any other textual key, including an empty string).

The `AppServer.HandleSSE()` handler provides a default endpoint for sending data to connected clients.

Here's a simple example that creates an SSE hub, hooks up a default handler, and then runs a goroutine that publishes to 
the SSE channels. (Note that in this case we use the -- expensive -- `sse.SSEHub.WriteAll()` method to write to all 
connected clients, which is suitable for messages that need to go to everyone.)

~~~
server.AddSSEHub("test")
server.Router.HandlerFunc(http.MethodGet, "/app/sse", server.HandleSSE("test", 72*60))

go func() {
    ticker := time.NewTicker(2 * time.Second)
    incr := 0
    go func() {
        for {
            select {
            case _ = <-ticker.C:
                incr++
                msg := "<div hx-ext=\"sse\" sse-connect=\"/app/sse\" sse-swap=\"message\">Message Count: " + strconv.Itoa(incr) + "</div>"
                evt := sse.SSEEvent{
                    UserID:    "",
                    ID:        "",
                    EventType: "",
                    Data:      []string{msg},
                    Retry:     0,
                }
                fmt.Println(evt.Dispatch())
                server.SSEHubs["test"].WriteAll(evt)
            }
        }
    }()
}()
~~~