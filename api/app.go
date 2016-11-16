package api

import (
	"fmt"
	"os"
	"strings"

	mgo "gopkg.in/mgo.v2"

	raven "github.com/getsentry/raven-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/fasthttp"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	newrelic "github.com/newrelic/go-agent"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
	fh "github.com/valyala/fasthttp"
)

// App is a struct that represents a Donations Api APP
type App struct {
	Debug        bool
	Port         int
	Host         string
	ConfigPath   string
	App          *echo.Echo
	Engine       engine.Server
	Config       *viper.Viper
	MongoDb      *mgo.Database
	MongoSession *mgo.Session
	Logger       zap.Logger
	Background   bool
	Fast         bool
	Client       *fh.Client
	NewRelic     newrelic.Application
}

// GetApp returns a new Donations Application
func GetApp(host string, port int, configPath string, debug bool, logger zap.Logger, background bool, fast bool) (*App, error) {
	app := &App{
		Host:       host,
		Port:       port,
		ConfigPath: configPath,
		Config:     viper.New(),
		Debug:      debug,
		Logger:     logger,
		Background: background,
		Fast:       fast,
	}
	err := app.Configure()
	if err != nil {
		return nil, err
	}
	return app, nil
}

// Configure instantiates the required dependencies for Donations Application
func (app *App) Configure() error {
	app.setConfigurationDefaults()

	err := app.loadConfiguration()
	if err != nil {
		return err
	}
	app.configureSentry()
	app.configureNewRelic()

	err = app.configureApplication()
	if err != nil {
		app.Logger.Error("Failed to configure application.", zap.Error(err))
		return err
	}

	app.Client = &fh.Client{}

	return nil
}

func (app *App) configureSentry() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureSentry"),
	)
	sentryURL := app.Config.GetString("sentry.url")
	l.Info("Configuring sentry...", zap.String("sentryURL", sentryURL))
	raven.SetDSN(sentryURL)
}

func (app *App) configureNewRelic() error {
	newRelicKey := app.Config.GetString("newrelic.key")
	appName := app.Config.GetString("newrelic.appName")
	if appName == "" {
		appName = "Khan"
	}

	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("appName", appName),
		zap.String("operation", "configureNewRelic"),
	)

	config := newrelic.NewConfig(appName, newRelicKey)
	if newRelicKey == "" {
		l.Info("New Relic is not enabled..")
		config.Enabled = false
	}
	nr, err := newrelic.NewApplication(config)
	if err != nil {
		l.Error("Failed to initialize New Relic.", zap.Error(err))
		return err
	}

	app.NewRelic = nr
	l.Info("Initialized New Relic successfully.")

	return nil
}

func (app *App) setConfigurationDefaults() {
	app.Config.SetDefault("healthcheck.workingText", "WORKING")
	app.Config.SetDefault("api.maxReadBufferSize", 32000)

	app.Config.SetDefault("mongo.host", "localhost")
	app.Config.SetDefault("mongo.port", 27017)
	app.Config.SetDefault("mongo.user", "")
	app.Config.SetDefault("mongo.password", "")
	app.Config.SetDefault("mongo.db", "donations")
}

func (app *App) loadConfiguration() error {
	app.Config.SetConfigFile(app.ConfigPath)
	app.Config.SetEnvPrefix("donations")
	app.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	app.Config.AutomaticEnv()

	if err := app.Config.ReadInConfig(); err == nil {
		app.Logger.Info("Loaded config file.", zap.String("configFile", app.Config.ConfigFileUsed()))
	} else {
		return fmt.Errorf("Could not load configuration file from: %s", app.ConfigPath)
	}

	return nil
}

//OnErrorHandler handles application panics
func (app *App) OnErrorHandler(err error, stack []byte) {
	app.Logger.Error(
		"Panic occurred.",
		zap.String("operation", "OnErrorHandler"),
		zap.Error(err),
		zap.String("stack", string(stack)),
	)

	var e error
	switch err.(type) {
	case error:
		e = err.(error)
	default:
		e = fmt.Errorf("%v", err)
	}

	tags := map[string]string{
		"source": "app",
		"type":   "panic",
	}
	raven.CaptureError(e, tags)
}

func (app *App) configureApplication() error {
	l := app.Logger.With(
		zap.String("operation", "configureApplication"),
	)

	l.Debug("Configuring Application...")
	addr := fmt.Sprintf("%s:%d", app.Host, app.Port)
	if app.Fast {
		app.Engine = fasthttp.New(addr)
	} else {
		app.Engine = standard.New(addr)
	}
	app.App = echo.New()
	a := app.App

	_, w, _ := os.Pipe()
	a.SetLogOutput(w)

	basicAuthUser := app.Config.GetString("api.basicAuth.user")
	if basicAuthUser != "" {
		basicAuthPass := app.Config.GetString("api.basicAuth.pass")

		a.Use(middleware.BasicAuth(func(username, password string) bool {
			return username == basicAuthUser && password == basicAuthPass
		}))
	}

	//NewRelicMiddleware has to stand out from all others
	a.Use(NewNewRelicMiddleware(app, app.Logger).Serve)

	a.Use(NewRecoveryMiddleware(app.OnErrorHandler).Serve)
	a.Use(NewVersionMiddleware().Serve)
	a.Use(NewSentryMiddleware(app).Serve)
	a.Use(NewLoggerMiddleware(app.Logger).Serve)
	a.Use(NewBodyExtractionMiddleware().Serve)

	a.Get("/healthcheck", HealthCheckHandler(app))

	a.Post("/games", CreateGameHandler(app))
	a.Put("/games/:gameID", UpdateGameHandler(app))

	app.configureMongoDB()

	l.Debug("Application configured successfully.")

	return nil
}

func (app *App) configureMongoDB() error {
	l := app.Logger.With(
		zap.String("operation", "configureMongoDB"),
		zap.String("mongo.host", app.Config.GetString("mongo.host")),
		zap.Int("mongo.port", app.Config.GetInt("mongo.port")),
	)
	session, err := mgo.Dial(fmt.Sprintf(
		"%s:%d",
		app.Config.GetString("mongo.host"),
		app.Config.GetInt("mongo.port"),
	))
	if err != nil {
		l.Fatal("Could not connect to MongoDb.", zap.Error(err))
	}
	session.SetMode(mgo.Monotonic, true)
	db := session.DB(app.Config.GetString("mongo.db"))
	app.MongoSession = session
	app.MongoDb = db
	return nil
}

// Start starts listening for web requests at specified host and port
func (app *App) Start() {
	l := app.Logger.With(
		zap.String("operation", "Start"),
	)

	l.Info("Starting Donations...", zap.String("host", app.Host), zap.Int("port", app.Port))
	if app.Background {
		go func() {
			app.App.Run(app.Engine)
		}()
	} else {
		app.App.Run(app.Engine)
	}
}

//Stop app running routines
func (app *App) Stop() {
	app.MongoSession.Close()
}
