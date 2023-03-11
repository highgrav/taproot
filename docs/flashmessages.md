# Flash Messages

Taproot has the ability to send "flash" messages back to a client. Flash messages can be stored in an `http.Request` 
object context or in the client's session. The `HandleFlashResponses()` middleware finds any existing flash 
messages and adds an `x-flash-msg` header to the HTTP response. Clients are expected to handle flash message headers; 
a message may be appended to more than one HTTP response, and messages are not guaranteed to be delivered if the client 
does not send a request.