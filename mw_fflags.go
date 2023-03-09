package taproot

import "net/http"

type FeatureFlagTargets map[string]map[string]http.HandlerFunc

func (srv *AppServer) HandleFflag(rules FeatureFlagTargets, defaultHandler http.HandlerFunc) http.HandlerFunc {

	return defaultHandler
}
