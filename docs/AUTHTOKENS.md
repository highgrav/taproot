# Authentication Tokens

Taproot includes a simple facility for encrypting tokens that can be included in headers or cookies. 
The `authn.AuthSignerManager` is responsible for rotating `AuthSigners` that AES-encrypt tokens. 
When creating a new server, you will need to pass a function in that can generate a string ID and []byte secret for AuthSigners;
if running a single node (or using session-aware proxies), you can use the default `authtoken.DefaultAuthSecretRotator` function, 
which just generates a random ID and secret. When using multiple servers, you'll likely want to synchronize secrets.


`HandleHeaderSession()` and `HandleCookieSession()` are sample middlewares that demonstrate using tokens in 
action. 