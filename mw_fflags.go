package taproot

import (
	"context"
	"github.com/highgrav/taproot/v1/authn"
	"github.com/highgrav/taproot/v1/common"
	"github.com/highgrav/taproot/v1/logging"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"net/http"
)

func (srv *AppServer) HandleFeatureFlags(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		flagKey, ok := ctx.Value(HTTP_CONTEXT_FFLAG_KEY).(string)
		if !ok || flagKey == "" {
			user, ok := ctx.Value(HTTP_CONTEXT_USER_KEY).(authn.User)
			if !ok {
				logging.LogToDeck("error", "FFLAG\terror\tcould not get user from context")
				next.ServeHTTP(w, r)
				return
			}
			if user.UserID == "" {
				sessKey, ok := ctx.Value(HTTP_CONTEXT_SESSION_KEY).(string)
				if !ok || sessKey == "" {
					sessKey = common.CreateRandString(24)
				}
				ffuser.NewAnonymousUser(sessKey)
				ctx = context.WithValue(ctx, HTTP_CONTEXT_FFLAG_KEY, sessKey)
			} else {
				ffuserbuilder := ffuser.NewUserBuilder(user.UserID)
				ffuserbuilder.AddCustom("userId", user.UserID)
				ffuserbuilder.AddCustom("realmId", user.RealmID)
				ffuserbuilder.AddCustom("domainId", user.DomainID)
				ffuserbuilder.AddCustom("userName", user.Username)
				ffuserbuilder.Build()
				ctx = context.WithValue(ctx, HTTP_CONTEXT_FFLAG_KEY, user.UserID)
			}
		} else {
			// check to see if we need to transition to a new ID
			sessKey, ok := ctx.Value(HTTP_CONTEXT_SESSION_KEY).(string)
			if !ok || sessKey == "" {
				next.ServeHTTP(w, r)
				return
			}
			user, ok := ctx.Value(HTTP_CONTEXT_USER_KEY).(authn.User)

			if !ok || user.UserID == "" {
				// user doesn't exist, but we do have a new session
				if flagKey != sessKey {
					ffuser.NewAnonymousUser(sessKey)
					ctx = context.WithValue(ctx, HTTP_CONTEXT_FFLAG_KEY, sessKey)
				}
			} else {
				if flagKey != user.UserID {
					// we have a user, so let's replace their flags
					logging.LogToDeck("info", "FFLAG\tinfo\tupgrading flags from "+flagKey+" for user "+user.UserID)
					ffuserbuilder := ffuser.NewUserBuilder(user.UserID)
					ffuserbuilder.AddCustom("userId", user.UserID)
					ffuserbuilder.AddCustom("realmId", user.RealmID)
					ffuserbuilder.AddCustom("domainId", user.DomainID)
					ffuserbuilder.AddCustom("userName", user.Username)
					ffuserbuilder.Build()
					ctx = context.WithValue(ctx, HTTP_CONTEXT_FFLAG_KEY, user.UserID)
				}
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
