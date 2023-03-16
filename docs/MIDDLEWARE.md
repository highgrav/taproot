# Middleware

### Example
The following is an example of adding some standard middleware into the chain. Note that request logging is not automatically 
started, and needs to be manually added into the chain.
~~~
server.AddMiddleware(myWebApp.HandleUserInjection)
server.AddMiddleware(server.HandleStaticFiles)
server.AddMiddleware(server.HandleLogging) // Only log non-static file requests
~~~