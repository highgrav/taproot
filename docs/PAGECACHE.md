# Page Caching

Taproot has a simple page caching capability for JS and JSML pages. By passing a positive integer into 
`server.HandleScript()` you tell Taproot to cache the page results for that many seconds:
~~~
server.Handler(http.MethodGet, "/", server.HandleScript("views/pages/index.jsml", 300, nil))
~~~
