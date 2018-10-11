package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/alexei-led/pumba/pkg/chaos"
)

const (
	kubeInterfaceKey = "kube.interface"
)

type kubeContext struct {
	context context.Context
}

// NewKubeCLICommand initialize kubernetes main command and bind it to the kubeContext
func NewKubeCLICommand(ctx context.Context) *cli.Command {
	cmdContext := &kubeContext{context: ctx}
	return &cli.Command{
		Name:    "kubernetes",
		Aliases: []string{"kube", "k8s"},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "context, c",
				Usage: "the name of the kubeconfig context to use",
				Value: "default",
			},
			cli.StringFlag{
				Name:  "kubeconfig",
				Usage: "path to the kubeconfig file to use for Kubernetes API requests",
				Value: "~/.kube/config",
			},
			cli.StringFlag{
				Name:  "duration, d",
				Usage: "kubernetes chaos duration; should be smaller than recurrent interval; use with optional unit suffix: 'ms/s/m/h'",
			},
		},
		Subcommands: []cli.Command{
			*NewScaleToZeroCLICommand(ctx),
		},
		Usage:       "chaos testing for Kubernetes",
		ArgsUsage:   fmt.Sprintf("services/pods/deployments: name/label, list of names/labels, or RE2 regex if prefixed with %q", chaos.Re2Prefix),
		Description: "tolerate random failures for Kubernetes resources: stability, network,  performance, resource limits and configuration issues",
		Before:      cmdContext.before,
	}
}

// Before any kubernetes sub-command runs
func (cmd *kubeContext) before(c *cli.Context) error {
	// get kubernetes context
	// kubeContext := c.String("context")
	// kubernetes config file
	kubeconfig := c.String("kubeconfig")
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}
	// create the clientset
	var kubeInterface kubernetes.Interface
	kubeInterface, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	// save kubernetes Interface in App metadata
	c.App.Metadata[kubeInterfaceKey] = kubeInterface
	return nil
}
