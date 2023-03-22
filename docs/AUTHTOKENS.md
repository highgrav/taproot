# Authentication Tokens

Taproot includes a simple facility for encrypting tokens that can be included in headers or cookies. 
The `authn.AuthSignerManager` is responsible for rotating `AuthSigners` that AES-encrypt tokens.



`HandleHeaderSession()` and `HandleCookieSession()` are sample middlewares that demonstrate using tokens in 
action. 