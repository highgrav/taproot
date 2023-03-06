package taproot

import (
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"sync"
	"time"
)

// TODO -- HttpServer config
func (srv *AppServer) HandleGlobalRateLimit(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(srv.Config.HttpServer.GlobalRateLimits.RequestsPerSecond), srv.Config.HttpServer.GlobalRateLimits.BurstableRequests)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			srv.RateLimitExceededResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (srv *AppServer) HandleIPRateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
		exempted bool
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)
	// cleanup client cache
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > time.Second*time.Duration(srv.Config.HttpServer.IpRateLimits.SweepClientCacheInSeconds) {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipstr, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			srv.ServerErrorResponse(w, r)
			return
		}
		mu.Lock()
		if _, exists := clients[ipstr]; !exists {
			ip := net.ParseIP(ipstr)
			if ip == nil {
				srv.ServerErrorResponse(w, r)
				return
			}
			for _, exempt := range srv.Config.HttpServer.IpRateLimits.ExemptNets {
				if exempt.Contains(ip) {
					clients[ipstr] = &client{
						lastSeen: time.Now(),
						exempted: true,
					}
					break
				}
			}
			if _, nowexists := clients[ipstr]; !nowexists {
				clients[ipstr] = &client{
					limiter:  rate.NewLimiter(rate.Limit(srv.Config.HttpServer.IpRateLimits.RequestsPerSecond), srv.Config.HttpServer.IpRateLimits.BurstableRequests),
					lastSeen: time.Now(),
					exempted: false,
				}
			}
		}
		clients[ipstr].lastSeen = time.Now()
		if !clients[ipstr].exempted && !clients[ipstr].limiter.Allow() {
			mu.Unlock()
			srv.RateLimitExceededResponse(w, r)
			return
		}
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
