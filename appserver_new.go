package taproot

import (
	"database/sql"
	"fmt"
	"github.com/google/deck"
	"github.com/google/deck/backends/logger"
	"github.com/jpillora/ipfilter"
	"github.com/julienschmidt/httprouter"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"highgrav/taproot/v1/acacia"
	"highgrav/taproot/v1/authn"
	"highgrav/taproot/v1/jsrun"
	"net/http"
	"os"
	"time"
)

func New(userStore authn.IUserStore, fflagretriever retriever.Retriever, cfgDirs []string) *AppServer {
	cfg, err := loadConfig(cfgDirs)
	if err != nil {
		panic(err)
	}

	svr := NewWithConfig(userStore, fflagretriever, cfg)
	return svr
}

func NewWithConfig(userStore authn.IUserStore, fflagretriever retriever.Retriever, cfg ServerConfig) *AppServer {
	// set up logging (we use stdout until the server is up and running)
	deck.Add(logger.Init(os.Stdout, 0))

	s := &AppServer{}
	s.Config = cfg
	s.users = userStore
	s.DBs = make(map[string]*sql.DB)
	s.Middleware = make([]MiddlewareFunc, 0)
	s.jsinjections = make([]jsrun.InjectorFunc, 0)

	// Set up IP filter
	// TODO
	s.httpIpFilter = ipfilter.New(ipfilter.Options{})

	// Set up our feature flags
	s.fflags = fflagretriever
	ffclient.Init(ffclient.Config{
		PollingInterval: time.Duration(s.Config.Flags.PollingInterval) * time.Second,
		Environment:     s.Config.Flags.Environment,
		Retriever:       s.fflags,
		Notifiers:       nil,
		FileFormat:      "",
		DataExporter:    ffclient.DataExporter{},
		Offline:         s.Config.Flags.Offline,
	})

	// Set up our security policy authorizer
	sa, err := acacia.New(cfg.SecurityPolicyDir)
	if err != nil {
		deck.Fatal(err.Error())
		os.Exit(-1)
	}
	s.authz = sa

	// set up our JS manager
	js, err := jsrun.New(cfg.ScriptFilePath)
	if err != nil {
		deck.Fatal(err.Error())
		os.Exit(-1)
	}
	s.js = js

	if s.Config.UseJSML {
		err = s.compileJSMLFiles(s.Config.JSMLFilePath, s.Config.JSMLCompiledFilePath)
		if err != nil {
			deck.Fatal(err.Error())
			os.Exit(-1)
		}
	}

	s.Router = httprouter.New()
	s.Router.SaveMatchedRoutePath = true // necessary to get the matched path back for Acacia authz
	s.Server = &WebServer{}
	s.Server.Server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HttpServer.ServerName, cfg.HttpServer.Port),
		Handler:      s.Router,
		IdleTimeout:  time.Duration(cfg.HttpServer.Timeouts.Idle) * time.Second,
		ReadTimeout:  time.Duration(cfg.HttpServer.Timeouts.Read) * time.Second,
		WriteTimeout: time.Duration(cfg.HttpServer.Timeouts.Write) * time.Second,
	}

	return s
}
