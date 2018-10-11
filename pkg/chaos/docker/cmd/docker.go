package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/urfave/cli"

	"github.com/alexei-led/pumba/pkg/chaos"
	netemCmd "github.com/alexei-led/pumba/pkg/chaos/netem/cmd"
	"github.com/alexei-led/pumba/pkg/container"
)

const (
	// Re2Prefix re2 regexp string prefix
	Re2Prefix = "re2:"
)

type dockerContext struct {
	context context.Context
}

// NewDockerCLICommand initialize docker main command and bind it to the Docker host
func NewDockerCLICommand(ctx context.Context) *cli.Command {
	rootCertPath := "/etc/ssl/docker"
	if os.Getenv("DOCKER_CERT_PATH") != "" {
		rootCertPath = os.Getenv("DOCKER_CERT_PATH")
	}
	cmdContext := &dockerContext{context: ctx}
	return &cli.Command{
		Name:    "docker",
		Aliases: []string{"d"},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "host, H",
				Usage:  "daemon socket to connect to",
				Value:  "unix:///var/run/docker.sock",
				EnvVar: "DOCKER_HOST",
			},
			cli.BoolFlag{
				Name:  "tls",
				Usage: "use TLS; implied by --tlsverify",
			},
			cli.BoolFlag{
				Name:   "tlsverify",
				Usage:  "use TLS and verify the remote",
				EnvVar: "DOCKER_TLS_VERIFY",
			},
			cli.StringFlag{
				Name:  "tlscacert",
				Usage: "trust certs signed only by this CA",
				Value: fmt.Sprintf("%s/ca.pem", rootCertPath),
			},
			cli.StringFlag{
				Name:  "tlscert",
				Usage: "client certificate for TLS authentication",
				Value: fmt.Sprintf("%s/cert.pem", rootCertPath),
			},
			cli.StringFlag{
				Name:  "tlskey",
				Usage: "client key for TLS authentication",
				Value: fmt.Sprintf("%s/key.pem", rootCertPath),
			},
		},
		Subcommands: []cli.Command{
			*NewKillCLICommand(ctx),
			*NewStopCLICommand(ctx),
			*NewPauseCLICommand(ctx),
			*NewRemoveCLICommand(ctx),
			*netemCmd.NewNetemCLICommand(ctx),
		},
		Usage:       "chaos testing for Docker",
		ArgsUsage:   fmt.Sprintf("services/pods/deployments: name/label, list of names/labels, or RE2 regex if prefixed with %q", chaos.Re2Prefix),
		Description: "tolerate random failures for Docker: process, network and performance",
		Before:      cmdContext.before,
	}
}

// Before any docker sub-command runs
func (cmd *dockerContext) before(c *cli.Context) error {
	/// Set-up container client
	tls, err := tlsConfig(c)
	if err != nil {
		return err
	}
	// create new Docker client
	chaos.DockerClient = container.NewClient(c.String("host"), tls)
	return nil
}

// tlsConfig translates the command-line options into a tls.Config struct
func tlsConfig(c *cli.Context) (*tls.Config, error) {
	var tlsConfig *tls.Config
	var err error
	caCertFlag := c.String("tlscacert")
	certFlag := c.String("tlscert")
	keyFlag := c.String("tlskey")

	if c.Bool("tls") || c.Bool("tlsverify") {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: !c.Bool("tlsverify"),
		}

		// Load CA cert
		if caCertFlag != "" {
			var caCert []byte
			if strings.HasPrefix(caCertFlag, "/") {
				caCert, err = ioutil.ReadFile(caCertFlag)
				if err != nil {
					return nil, err
				}
			} else {
				caCert = []byte(caCertFlag)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}

		// Load client certificate
		if certFlag != "" && keyFlag != "" {
			var cert tls.Certificate
			if strings.HasPrefix(certFlag, "/") && strings.HasPrefix(keyFlag, "/") {
				cert, err = tls.LoadX509KeyPair(certFlag, keyFlag)
				if err != nil {
					return nil, err
				}
			} else {
				cert, err = tls.X509KeyPair([]byte(certFlag), []byte(keyFlag))
				if err != nil {
					return nil, err
				}
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}
	return tlsConfig, nil
}
