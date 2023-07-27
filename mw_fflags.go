package taproot

import (
	"github.com/highgrav/taproot/authn"
	"github.com/highgrav/taproot/constants"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"net/http"
)

func (srv *AppServer) HandleFeatureFlags(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sessKey, ok := ctx.Value(constants.HTTP_CONTEXT_SESSION_KEY).(string)
		if !ok || sessKey == "" {
			next.ServeHTTP(w, r)
			return
		}

		user, ok := ctx.Value(constants.HTTP_CONTEXT_USER_KEY).(authn.User)
		if !ok || user.UserID == "" {
			ffuser.NewAnonymousUser(sessKey)
		} else {
			// we have a user, so let's replace their flags
			ffuserbuilder := ffuser.NewUserBuilder(user.UserID)
			ffuserbuilder.AddCustom("userId", user.UserID)
			ffuserbuilder.AddCustom("realmId", user.RealmID)
			ffuserbuilder.AddCustom("domainId", user.DomainID)
			ffuserbuilder.AddCustom("userName", user.Username)
			ffuserbuilder.Build()
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
