// donations
// https://github.com/topfreegames/donations
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/topfreegames/donations/api"
	"github.com/topfreegames/donations/log"
	"github.com/uber-go/zap"
)

var host string
var port int
var debug bool
var quiet bool
var fast bool
var configFile string

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the donations API server",
	Long: `Starts donations server with the specified arguments. You can use
environment variables to override configuration keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		ll := zap.InfoLevel
		if debug {
			ll = zap.DebugLevel
		}
		if quiet {
			ll = zap.ErrorLevel
		}
		l := zap.New(
			zap.NewJSONEncoder(), // drop timestamps in tests
			ll,
		)

		cmdL := l.With(
			zap.String("source", "startCmd"),
			zap.String("operation", "Run"),
			zap.String("host", host),
			zap.Int("port", port),
			zap.Bool("debug", debug),
		)

		log.D(cmdL, "Creating application...")
		app, err := api.GetApp(
			host,
			port,
			configFile,
			debug,
			l,
			false,
			fast,
		)

		if err != nil {
			log.E(cmdL, "Application failed to start.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			os.Exit(1)
		}
		log.D(cmdL, "Application created successfully.")

		log.I(cmdL, fmt.Sprintf("Application started successfully at %s:%d", host, port))
		app.Start()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVarP(&host, "bind", "b", "0.0.0.0", "Host to bind donations to")
	startCmd.Flags().IntVarP(&port, "port", "p", 8888, "Port to bind donations to")
	startCmd.Flags().StringVarP(&configFile, "config", "c", "./config/default.yaml", "Configuration path")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
	startCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode (log level error)")
	startCmd.Flags().BoolVarP(&fast, "fast", "f", false, "Use FastHTTP as the Engine. If false, the net/http engine will be used")
}
