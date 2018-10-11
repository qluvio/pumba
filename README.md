# Pumba: Chaos testing tool for Docker

[![Join the chat at https://gitter.im/pumba-chaos/Lobby](https://badges.gitter.im/pumba-chaos/Lobby.svg)](https://gitter.im/pumba-chaos/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

[![Build Status](https://travis-ci.org/alexei-led/pumba.svg?branch=master)](https://travis-ci.org/alexei-led/pumba) [![Codefresh build status]( https://g.codefresh.io/api/badges/build?repoOwner=alexei-led&repoName=pumba&branch=master&pipelineName=pumba&accountName=codefresh-inc&type=cf-1)]( https://g.codefresh.io/repositories/alexei-led/pumba/builds?filter=trigger:build;branch:master;service:5a9d1dac81caf90001f95f9d~pumba) [![Go Report Card](https://goreportcard.com/badge/github.com/alexei-led/pumba)](https://goreportcard.com/report/github.com/alexei-led/pumba) [![codecov](https://codecov.io/gh/alexei-led/pumba/branch/master/graph/badge.svg)](https://codecov.io/gh/alexei-led/pumba)

[![](https://badge.imagelayers.io/gaiaadm/pumba:master.svg)](https://imagelayers.io/?images=gaiaadm/pumba:master)  [![](https://images.microbadger.com/badges/image/gaiaadm/pumba.svg)](http://microbadger.com/images/gaiaadm/pumba) [![](https://images.microbadger.com/badges/version/gaiaadm/pumba.svg)](http://microbadger.com/images/gaiaadm/pumba) [![](https://images.microbadger.com/badges/commit/gaiaadm/pumba.svg)](http://microbadger.com/images/gaiaadm/pumba) [![Anchore Image Overview](https://anchore.io/service/badges/image/77101bee4abccf2db02413002f25930b73bc6f5fea187b1b5ab1f0b538c1ba7a)](https://anchore.io/image/dockerhub/77101bee4abccf2db02413002f25930b73bc6f5fea187b1b5ab1f0b538c1ba7a?repo=gaiaadm%2Fpumba&tag=latest#overview)

## Logo

![pumba](docs/img/pumba_logo.png)

## Demo

[![asciicast](https://asciinema.org/a/82428.png)](https://asciinema.org/a/82428)

## Usage

You can download Pumba binary for your OS from [release](https://github.com/alexei-led/pumba/releases) page.

```text
$ pumba --help

NAME:
   Pumba - is a resilience/chaos testing tool, for Docker and Kubernetes

USAGE:
   pumba [global options] command [command options] [arguments...]

VERSION:
   [VERSION](./blob/master/VERSION) - `git rev-parse HEAD --short` and `build time`

COMMANDS:
     docker, d              chaos testing for Docker
     kubernetes, kube, k8s  chaos testing for Kubernetes
     help, h                Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log-level value, -l value  set log level (debug, info, warning(*), error, fatal, panic) (default: "warning") [$LOG_LEVEL]
   --json, -j                   produce log in JSON format: Logstash and Splunk friendly [$LOG_JSON]
   --slackhook value            web hook url; send Pumba log events to Slack
   --slackchannel value         Slack channel (default #pumba) (default: "#pumba")
   --interval value, -i value   recurrent interval for chaos command; use with optional unit suffix: 'ms/s/m/h'
   --random, -r                 randomly select single matching target from list of targets
   --dry-run                    dry run does not create chaos, only logs planned chaos commands [$DRY-RUN]
   --help, -h                   show help
   --version, -v                print the version

```

##### Examples

```text
# add 3 seconds delay for all outgoing packets on device `eth0` (default) of `mydb` Docker container for 5 minutes

$ pumba docker netem --duration 5m delay --time 3000 mydb
```

```text
# add a delay of 3000ms ± 30ms, with the next random element depending 20% on the last one,
# for all outgoing packets on device `eth1` of all Docker container, with name start with `hp`
# for 5 minutes

$ pumba docker netem --duration 5m --interface eth1 delay \
      --time 3000 \
      --jitter 30 \
      --correlation 20 \
    re2:^hp
```

```text
# add a delay of 3000ms ± 40ms, where variation in delay is described by `normal` distribution,
# for all outgoing packets on device `eth0` of randomly chosen Docker container from the list
# for 5 minutes

$ pumba --random docker netem --duration 5m \
    delay \
      --time 3000 \
      --jitter 40 \
      --distribution normal \
    container1 container2 container3
```

```text
# Corrupt 10% of the packets from the `mydb` Docker container for 5 minutes

$ pumba docker netem --duration 5m corrupt --percent 10 mydb
```

##### `tc` tool

Pumba uses `tc` Linux tool for network emulation. You have two options:

1. Make sure that container, you want to disturb, has `tc` tool available and properly installed (install `iproute2` package)
2. Use `--tc-image` option, with any `netem` command, to specify external Docker image with `tc` tool available. Pumba will create a new container from this image, adding `NET_ADMIN` capability to it and reusing target container network stack. You can try to use [gaiadocker/iproute2](https://hub.docker.com/r/gaiadocker/iproute2/) image (it's just Alpine Linux 3.3 with `iproute2` package installed)

**Note:** For Alpine Linux based image, you need to install `iproute2` package and also to create a symlink pointing to distribution files `ln -s /usr/lib/tc /lib/tc`.

### Running inside Docker container

If you choose to use Pumba Docker [image](https://hub.docker.com/r/gaiaadm/pumba/) on Linux, use the following command:

```text
# once in a 10 seconds, try to kill (with `SIGTERM` signal) all containers named **hp(something)**
# on same Docker host, where Pumba container is running

$ docker run -d -v /var/run/docker.sock:/var/run/docker.sock gaiaadm/pumba --interval 10s docker kill --signal SIGTERM ^hp
```

**Note:** from version `0.6` Pumba Docker image is a `scratch` Docker image, that contains only single `pumba` binary file and `ENTRYPOINT` set to the `pumba` command.

**Note:** For Windows and OS X you will need to use `--host` argument, since there is no unix socket `/var/run/docker.sock` to mount.

### Chaos Testing for Kubernetes

Pumba uses Kubernetes API to inject chaos into Kubernetes application.

## Build instructions

You can build Pumba with or without Go installed on your machine.

### Build using local Go environment

In order to build Pumba, you need to have Go 1.6+ setup on your machine.

Here is the approximate list of commands you will need to run:

```sh
# create required folder
cd $GOPATH
mkdir github.com/alexei-led && cd github.com/alexei-led

# clone pumba
git clone git@github.com:alexei-led/pumba.git
cd pumba

# build pumba binary
./hack/build.sh

# run tests and create HTML coverage report
CGO_ENABLED=1 ./hack/test.sh --html

# create pumba binaries for multiple platforms
./hack/xbuild.sh
```

### Build using Docker

You do not have to install and configure Go in order to build and test Pumba project. Pumba uses Docker multistage build to create final tiny Docker image.

First of all clone Pumba git repository:

```sh
git clone git@github.com:alexei-led/pumba.git
cd pumba
```

Now create a new Pumba Docker image.

```sh
docker build -t pumba -f Dockerfile .
```

## License

Code is under the [Apache License v2](https://www.apache.org/licenses/LICENSE-2.0.txt).
