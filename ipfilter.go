package taproot

import "github.com/jpillora/ipfilter"

func newIpFilter(cfg IPFilterConfig) *ipfilter.IPFilter {
	f := ipfilter.New(ipfilter.Options{
		AllowedIPs:       cfg.AllowedCidrs,
		BlockedIPs:       cfg.BlockedCidrs,
		AllowedCountries: cfg.AllowedCountries,
		BlockedCountries: cfg.BlockedCountries,
		BlockByDefault:   cfg.BlockByDefault,
		TrustProxy:       false,
	})
	return f
}
