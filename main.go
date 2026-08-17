/*
Chronokeep Desktop - Race Scoring Software
Copyright (C) 2026 James Sentinella

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"chronokeep/remote/handlers"
	"chronokeep/remote/util"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"golang.org/x/crypto/acme"
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

	log.Info("Setting up base middleware.")
	// Set up Recover and Logger middleware
	e.Use(middleware.Recover())
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogMethod:   true,
		LogLatency:  true,
		HandleError: true, // forwards error to the global error handler
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("latency", v.Latency.String()),
					slog.String("method", v.Method),
					slog.Int("status", v.Status),
					slog.String("uri", v.URI),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("latency", v.Latency.String()),
					slog.String("method", v.Method),
					slog.Int("status", v.Status),
					slog.String("uri", v.URI),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
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

	e.Any("/health", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	if config.Development {
		e.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
			Handler: func(c *echo.Context, req []byte, res []byte, err error) {
				log.Info("Request Log: ", string(req))
				log.Info("Request Header: ", c.Request().Header)
				log.Info("Response Log: ", string(res))
				log.Info("Response Error: ", err)
			},
			Skipper: healthEndpointSkipper,
		}))
	}
	if !config.AutoTLS {
		log.Info("Starting non https echo server.")
		log.Fatal(e.Start(":" + strconv.Itoa(config.Port)))
	} else {
		log.Info("Starting auto tls echo server.")
		// Set up auto tls manager - Cache certificates
		autoTLSManager := autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  autocert.DirCache("/var/www/.cache"),
		}
		// Set up TLS with auto certificate if not a debug environment.
		if !config.Development {
			autoTLSManager.HostPolicy = autocert.HostWhitelist(config.Domain)
		}
		e.Pre(middleware.HTTPSRedirect())
		s := http.Server{
			Addr:    ":" + strconv.Itoa(config.Port),
			Handler: e,
			TLSConfig: &tls.Config{
				GetCertificate: autoTLSManager.GetCertificate,
				NextProtos:     []string{acme.ALPNProto},
			},
		}
		log.Fatal(s.ListenAndServeTLS("", ""))
	}
}

func healthEndpointSkipper(c *echo.Context) bool {
	if c.Request().URL.Path == "/account/login" {
		return true
	}
	return strings.HasPrefix(c.Path(), "/health")
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

