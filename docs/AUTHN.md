# Users and Authentication


### Users and Middleware
You should always inject a user into your middleware chain, even for public pages. An empty user is treated as a public 
or anonymous user, so you can apply Acacia scripts to that user by testing for a blank user ID.

