# Sessions
Taproot uses a session abstraction mechanism that is mostly signature-compatible with the SCS library, in order to allow 
the use of its existing integrations. Taproot will attempt to rehydrate session User data from either a cookie or an 
X-Session header.

For API-based logins, we recommend using the X-Session header via `AddSessionHeader(http.ResponseWriter, string)`

For web-based logins, we recommend using cookies through `AddSessionCookie(http.ResponseWriter, string)`.

You can mix and match cookie- and header-based sessions freely, even sending a header and a cookie back in the same login 
response.

### Example

API Header-Based Session Example:
~~~
_, sessionkey, err := wapp.Taproot.RegisterUser(authn.UserAuth{
			AuthType:        authn.AUTH_BASIC,
			Realm:           "my-realm",
			Provider:        "my-provider",
			UserIdentifier:  "test@example.com,
			PasswordOrToken: "test123,
			ResetToken:      "",
		})

		if err != nil {
            // handle error 
			return
		}
		/*
		    Writes an X-Session header back to the client with an encrypted session token.
		*/
		err = wapp.Taproot.AddSessionHeader(w, sessionkey)
		if err != nil {
		    // handle error
		}
~~~