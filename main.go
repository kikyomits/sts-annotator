package main

import (
	"fmt"
	"strings"

	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func setupLogger(config *Config) {
	var logLevel string
	if config.Server.Mode == zerolog.DebugLevel.String() {
		log.Debug().Msg("Start with debug mode")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		gin.SetMode(gin.ReleaseMode)
		logLevel = strings.ToLower(config.Server.Log.Level)

		if logLevel == zerolog.DebugLevel.String() {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else if logLevel == zerolog.InfoLevel.String() {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		} else if logLevel == zerolog.WarnLevel.String() {
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		} else if logLevel == zerolog.ErrorLevel.String() {
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			log.Warn().Msgf("Received invalid log level to Server.Log.Level: '%v'. Use default log level %s", logLevel, zerolog.InfoLevel.String())
			logLevel = zerolog.InfoLevel.String()
		}
	}
	log.Info().Msgf("Log Level: %s", logLevel)
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"
}

func setupRouter(config *Config) (router *gin.Engine) {

	c := newConstant()

	router = gin.New()

	router.Use(logger.SetLogger(logger.Config{
		//Logger:   &subLog,
		UTC:      true,
		SkipPath: []string{c.V1 + c.Healthz},
	}))

	ctrl := newController(config)

	v1 := router.Group(c.V1)
	v1.GET(c.Healthz, health)
	v1.POST(c.Annotator, ctrl.annotateStsPod)
	return
}

func main() {
	config := initConfig()
	setupLogger(config)

	router := setupRouter(config)
	err := router.RunTLS(
		fmt.Sprintf(":%v", config.Server.Port),
		config.Server.Tls.Cert,
		config.Server.Tls.Key)

	if err != nil {
		log.Fatal().Err(err)
	}
}
