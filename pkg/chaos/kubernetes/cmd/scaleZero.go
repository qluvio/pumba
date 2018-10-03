package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/alexei-led/pumba/pkg/chaos"
	"github.com/alexei-led/pumba/pkg/chaos/kubernetes"
)

// NewScaleToZeroCLICommand initialize patch command and bind it to the kubeContext
func NewScaleToZeroCLICommand(ctx context.Context) *cli.Command {
	cmdContext := &kubeContext{context: ctx}
	return &cli.Command{
		Name: "scale-to-zero",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "kind,k",
				Usage: "deployment(deploy)replicaset(rs)|relicationcontroller(rc)|statefulset",
				Value: "deployment",
			},
			cli.StringFlag{
				Name:  "namespace,n",
				Usage: "kubernetes namespace name",
				Value: "default",
			},
		},
		Usage:       "set number of replicas to zero",
		ArgsUsage:   fmt.Sprintf("name/label, list of names/labels, or RE2 regex if prefixed with %q", chaos.Re2Prefix),
		Description: "set number of replicas for specified resource to zero",
		Action:      cmdContext.scaleToZero,
	}
}

// scaleToZero Command
func (cmd *kubeContext) scaleToZero(c *cli.Context) error {
	// get random
	random := c.GlobalBool("random")
	// get dry-run mode
	dryRun := c.GlobalBool("dry-run")
	// get interval
	interval := c.GlobalString("interval")
	// get names or pattern
	names, pattern := chaos.GetNamesOrPattern(c)
	// get kind
	kind := c.String("kind")
	// get namespace
	namespace := c.String("namespace")
	// get duration from parent `kube` command
	duration := c.Parent().String("duration")
	// init scale to zero command
	scaleToZeroCommand, err := kubernetes.NewScaleToZeroCommand(c.App.Metadata[kubeInterfaceKey], kind, namespace, names, pattern, interval, duration, dryRun)
	if err != nil {
		return err
	}
	// run chaos command
	return chaos.RunChaosCommand(cmd.context, scaleToZeroCommand, interval, random)
}
