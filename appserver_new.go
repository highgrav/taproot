package taproot

import (
	"database/sql"
	"expvar"
	"fmt"
	"github.com/google/deck"
	"github.com/google/deck/backends/logger"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"highgrav/taproot/v1/acacia"
	"highgrav/taproot/v1/authn"
	"highgrav/taproot/v1/jsrun"
	"highgrav/taproot/v1/logging"
	"highgrav/taproot/v1/workers"
	"net/http"
	"os"
	"time"
)

/*
Creates a new AppServer; uses Viper to populate the config from YAML files.
Requires the user pass in an IUserStore and a fflag retriever, as well as
the directories to search for default config files.
*/
func New(userStore authn.IUserStore, fflagretriever retriever.Retriever, cfgDirs []string) *AppServer {
	cfg, err := loadConfig(cfgDirs)
	if err != nil {
		panic(err)
	}
	// TODO
	svr := NewWithConfig(userStore, fflagretriever, cfg)
	return svr
}

// Creates a new AppServer using a ServerConfig struct.
func NewWithConfig(userStore authn.IUserStore, fflagretriever retriever.Retriever, cfg ServerConfig) *AppServer {
	// set up logging (we use stdout until the server is up and running)
	deck.Add(logger.Init(os.Stdout, 0))

	s := &AppServer{}
	s.Config = cfg
	s.users = userStore
	s.DBs = make(map[string]*sql.DB)
	s.Middleware = make([]alice.Constructor, 0)
	s.jsinjections = make([]jsrun.InjectorFunc, 0)

	logging.LogToDeck("info", "Setting up async work hub")
	wh, err := workers.New(cfg.WorkHub.Name, cfg.WorkHub.StorageDir, cfg.WorkHub.SegmentSize)
	if err != nil {
		logging.LogToDeck("fatal", err.Error())
		panic(err)
	}
	s.WorkHub = wh

	// Set up IP filter
	logging.LogToDeck("info", "Setting up IP filtering")
	s.httpIpFilter = newIpFilter(cfg.HttpServer.IPFilter)

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

	// Set up stats
	s.stats = make(map[string]stats)
	s.globalStats = stats{
		requests:       expvar.NewInt("total requests received"),
		responses:      expvar.NewInt("total responses sent"),
		processingTime: expvar.NewInt("total processing time in microsecs"),
		responseCodes:  expvar.NewMap("total responses by HTTP code"),
	}

	// Set up our security policy authorizer
	sa := acacia.NewPolicyManager()
	s.Acacia = sa
	err = s.Acacia.LoadAllFrom(cfg.SecurityPolicyDir)
	if err != nil {
		logging.LogToDeck("fatal", err.Error())
		panic(err)
	}

	if s.Config.UseJSML {
		err = s.compileJSMLFiles(s.Config.JSMLFilePath, s.Config.JSMLCompiledFilePath)
		if err != nil {
			logging.LogToDeck("fatal", err.Error())
			os.Exit(-1)
		}
	}

	// set up our JS manager
	js, err := jsrun.New(cfg.ScriptFilePath)
	if err != nil {
		logging.LogToDeck("fatal", err.Error())
		os.Exit(-1)
	}
	s.js = js

	s.Router = httprouter.New()
	s.Router.SaveMatchedRoutePath = true // necessary to get the matched path back for Acacia Acacia
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
