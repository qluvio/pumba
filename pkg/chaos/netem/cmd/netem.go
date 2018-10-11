package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli"
)

const (
	// Re2Prefix re2 regexp string prefix
	Re2Prefix = "re2:"
	// DefaultInterface default network interface
	DefaultInterface = "eth0"
)

// NewNetemCLICommand initialize docker netem sub-commands
func NewNetemCLICommand(ctx context.Context) *cli.Command {
	return &cli.Command{
		Name: "netem",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "duration, d",
				Usage: "network emulation duration; should be smaller than recurrent interval; use with optional unit suffix: 'ms/s/m/h'",
			},
			cli.StringFlag{
				Name:  "interface, i",
				Usage: "network interface to apply delay on",
				Value: DefaultInterface,
			},
			cli.StringSliceFlag{
				Name:  "target, t",
				Usage: "target IP filter; supports multiple IPs",
			},
			cli.StringFlag{
				Name:  "tc-image",
				Usage: "Docker image with tc (iproute2 package); try 'gaiadocker/iproute2'",
			},
		},
		Usage:       "emulate the properties of wide area networks",
		ArgsUsage:   fmt.Sprintf("containers (name, list of names, or RE2 regex if prefixed with %q", Re2Prefix),
		Description: "delay, loss, duplicate and re-order (run 'netem') packets, and limit the bandwidth, to emulate different network problems",
		Subcommands: []cli.Command{
			*NewDelayCLICommand(ctx),
			*NewLossCLICommand(ctx),
			*NewLossStateCLICommand(ctx),
			*NewLossGECLICommand(ctx),
			*NewRateCLICommand(ctx),
			*NewDuplicateCLICommand(ctx),
			*NewCorruptCLICommand(ctx),
		},
	}
}
