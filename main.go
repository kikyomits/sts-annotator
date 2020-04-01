package main

import (
	"fmt"
	"strings"

	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var CONFIG = initConfig()

func init() {

	CONFIG := initConfig()
	var logLevel string
	if CONFIG.Server.Mode == "debug" {
		log.Debug().Msg("Start with debug mode")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		logLevel = "debug"
	} else {
		gin.SetMode(gin.ReleaseMode)
		logLevel = strings.ToLower(CONFIG.Server.Log.Level)

		if logLevel == "debug" {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else if logLevel == "info" {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		} else if logLevel == "warn" {
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		} else if logLevel == "error" {
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			logLevel = "info"
			log.Warn().Msgf("Received invalid log level to Server.Log.Level: '%v'. Set 'INFO' to log level.", logLevel)
		}
	}
	log.Info().Msgf("Log Level: %s", logLevel)
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"
}

func setupRouter() (router *gin.Engine) {
	c := newConstant()

	router = gin.New()

	router.Use(logger.SetLogger(logger.Config{
		//Logger:   &subLog,
		UTC:      true,
		SkipPath: []string{c.V1 + c.Healthz},
	}))

	v1 := router.Group(c.V1)
	v1.GET(c.Healthz, health)
	v1.POST(c.Annotator, annotateStsPod)
	return
}

func main() {
	router := setupRouter()
	router.RunTLS(
		fmt.Sprintf(":%v", CONFIG.Server.Port),
		CONFIG.Server.Tls.Cert,
		CONFIG.Server.Tls.Key)
}
