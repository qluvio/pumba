package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	dockerCmd "github.com/alexei-led/pumba/pkg/chaos/docker/cmd"
	kubeCmd "github.com/alexei-led/pumba/pkg/chaos/kubernetes/cmd"
	"github.com/alexei-led/pumba/pkg/logger"

	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli"

	"github.com/johntdyer/slackrus"
)

var (
	topContext context.Context
)

var (
	// Version that is passed on compile time through -ldflags
	Version = "dev.build"

	// GitCommit that is passed on compile time through -ldflags
	GitCommit = "none"

	// GitBranch that is passed on compile time through -ldflags
	GitBranch = "none"

	// BuildTime that is passed on compile time through -ldflags
	BuildTime = "none"

	// HumanVersion is a human readable app version
	HumanVersion = fmt.Sprintf("%s - %.7s (%s) %s", Version, GitCommit, GitBranch, BuildTime)
)

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func init() {
	// set log level
	log.SetLevel(log.WarnLevel)
	log.SetFormatter(&log.TextFormatter{})
	// handle termination signal
	topContext = handleSignals()
}

func main() {
	app := cli.NewApp()
	app.Name = "Pumba"
	app.Version = HumanVersion
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "Alexei Ledenev",
			Email: "alexei.led@gmail.com",
		},
	}
	app.EnableBashCompletion = true
	app.Usage = "is a resilience/chaos testing tool, for Docker and Kubernetes"
	app.Before = before
	app.Commands = []cli.Command{
		*dockerCmd.NewDockerCLICommand(topContext),
		*kubeCmd.NewKubeCLICommand(topContext),
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "log-level, l",
			Usage:  "set log level (debug, info, warning(*), error, fatal, panic)",
			Value:  "warning",
			EnvVar: "LOG_LEVEL",
		},
		cli.BoolFlag{
			Name:   "json, j",
			Usage:  "produce log in JSON format: Logstash and Splunk friendly",
			EnvVar: "LOG_JSON",
		},
		cli.StringFlag{
			Name:  "slackhook",
			Usage: "web hook url; send Pumba log events to Slack",
		},
		cli.StringFlag{
			Name:  "slackchannel",
			Usage: "Slack channel (default #pumba)",
			Value: "#pumba",
		},
		cli.StringFlag{
			Name:  "interval, i",
			Usage: "recurrent interval for chaos command; use with optional unit suffix: 'ms/s/m/h'",
		},
		cli.BoolFlag{
			Name:  "random, r",
			Usage: "randomly select single matching target from list of targets",
		},
		cli.BoolFlag{
			Name:   "dry-run",
			Usage:  "dry run does not create chaos, only logs planned chaos commands",
			EnvVar: "DRY-RUN",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func before(c *cli.Context) error {
	// set debug log level
	switch level := c.GlobalString("log-level"); level {
	case "debug", "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "info", "INFO":
		log.SetLevel(log.InfoLevel)
	case "warning", "WARNING":
		log.SetLevel(log.WarnLevel)
	case "error", "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "fatal", "FATAL":
		log.SetLevel(log.FatalLevel)
	case "panic", "PANIC":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}
	// set log formatter to JSON
	if c.GlobalBool("json") {
		log.SetFormatter(&log.JSONFormatter{})
	}
	// set Slack log channel
	if c.GlobalString("slackhook") != "" {
		log.AddHook(&slackrus.SlackrusHook{
			HookURL:        c.GlobalString("slackhook"),
			AcceptedLevels: slackrus.LevelThreshold(log.GetLevel()),
			Channel:        c.GlobalString("slackchannel"),
			IconEmoji:      ":boar:",
			Username:       "pumba_bot",
		})
	}
	// trace function calls
	traceHook := logger.NewHook()
	traceHook.AppName = "pumba"
	log.AddHook(traceHook)
	return nil
}

func handleSignals() context.Context {
	// Graceful shut-down on SIGINT/SIGTERM
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// create cancelable context
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()
		sid := <-sig
		log.Debugf("Received signal: %d", sid)
		log.Debug("Canceling running chaos commands ...")
		log.Debug("Gracefully exiting after some cleanup ...")
	}()

	return ctx
}
