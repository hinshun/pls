package rethinkdbcmd

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/hinshun/pls/docker/dockercli"
	"github.com/hinshun/pls/docker/mitmproxy"
	"github.com/palantir/stacktrace"
	"gopkg.in/urfave/cli.v2"
)

func CreateRethinkdbREPL(c *cli.Context) error {
	panic("unimplemented")

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return stacktrace.Propagate(err, "failed to create docker client from env: %s", err)
	}

	err = dockercli.LazyImageLoad(ctx, cli, mitmproxy.MITMProxyImageName)
	if err != nil {
		return stacktrace.Propagate(err, "failed to load mitmproxy image")
	}

	return nil
}
