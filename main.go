package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"chronokeep/remote/handlers"
	"chronokeep/remote/util"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/acme/autocert"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Starting remote.")
	config, err := util.GetConfig()
	if err != nil {
		log.Fatal("Failed to get configuration. ", err)
	}
	e := echo.New()
	e.Debug = config.Development

	log.Info("Setting up base middleware.")
	// Set up Recover and Logger middleware
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format:  "${method} | ${status} | ${uri} -> ${latency_human}\n",
		Skipper: healthEndpointSkipper,
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"*",
		},
	}))

	log.Info("Calling handler setup.")
	// Handlers has a setup function which sets up the database for use.
	err = handlers.Setup(config)
	defer handlers.Finalize()
	if err != nil {
		log.Fatalf("Error setting up database. %v", err)
	}
	log.Info("Binding ")
	// Set up API handlers.
	handler := handlers.Handler{}
	// Setup the Handler for validator
	handler.Setup()
	handler.Bind(e.Group(""))
	handler.BindRestricted(e.Group(""))

	e.Any("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	if config.Development {
		e.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
			Handler: func(c echo.Context, req []byte, res []byte) {
				log.Info("Request Log: ", string(req))
				log.Info("Request Header: ", c.Request().Header)
				//log.Info("Response Log: ", string(res))
			},
			Skipper: healthEndpointSkipper,
		}))
	}
	if !config.AutoTLS {
		log.Info("Starting non https echo server.")
		log.Fatal(e.Start(":" + strconv.Itoa(config.Port)))
	} else {
		log.Info("Starting auto tls echo server.")
		// Set up TLS with auto certificate if not a debug environment.
		if !config.Development {
			e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(config.Domain)
		}
		// Cache certificates
		e.AutoTLSManager.Prompt = autocert.AcceptTOS
		e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
		e.Pre(middleware.HTTPSRedirect())
		log.Fatal(e.StartAutoTLS(":" + strconv.Itoa(config.Port)))
	}
}

func healthEndpointSkipper(c echo.Context) bool {
	if c.Request().URL.Path == "/account/login" {
		return true
	}
	return strings.HasPrefix(c.Path(), "/health")
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}
