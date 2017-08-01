package command

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/hinshun/pls/docker/dockercli"
	"github.com/hinshun/pls/docker/mitmproxy"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"

	"gopkg.in/urfave/cli.v2"
)

func CreateMITMProxy(c *cli.Context) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return stacktrace.Propagate(err, "failed to create docker client from env: %s", err)
	}

	err = dockercli.LazyImageLoad(ctx, cli, mitmproxy.MITMProxyImageName)
	if err != nil {
		return stacktrace.Propagate(err, "failed to load mitmproxy image")
	}

	spec := mitmproxy.MITMProxySpec{
		Name: c.String("name"),
	}

	mitmProxy, err := mitmproxy.New(ctx, cli, spec)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create mitmproxy")
	}

	logrus.Infof("Created mitmproxy container '%s'", mitmProxy.Name)

	return nil
}