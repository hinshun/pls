package rethinkdbrepl

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/hinshun/pls/pkg/namegen"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	RethinkDBREPLImageName = "jlhawn/rethinkdb-repl"
	RethinkDBREPLPrefix    = "rethinkdb-repl"
)

type RethinkDBREPLSpec struct {
	Name          string
	ServerAddress string
	ClientPort    string
}

type RethinkDBREPL struct {
	Name string

	rootCTX context.Context
	rootCLI client.APIClient
}

func New(ctx context.Context, cli client.APIClient, spec RethinkDBREPLSpec) (*RethinkDBREPL, error) {
	replName := spec.Name
	if replName == "" {
		var err error
		replName, err = namegen.GetUnusedContainerName(ctx, cli, RethinkDBREPLPrefix)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate rethinkdb repl container name")
		}
	}

	cfg := &container.Config{
		Image: RethinkDBREPLImageName,
		Labels: map[string]string{
			"pls": RethinkDBREPLPrefix,
		},
		Cmd:         []string{spec.ServerAddress, spec.ClientPort},
		Tty:         true,
		AttachStdin: true,
		OpenStdin:   true,
	}
	hostCfg := &container.HostConfig{}
	netCfg := &network.NetworkingConfig{}

	logrus.Infof("Creating container '%s'", replName)
	createResp, err := cli.ContainerCreate(ctx, cfg, hostCfg, netCfg, replName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create rethinkdb repl container")
	}

	err = cli.ContainerStart(ctx, createResp.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to start rethinkdb repl container")
	}

	return &RethinkDBREPL{
		ID:   createResp.ID,
		Name: replName,

		rootCTX: ctx,
		rootCLI: cli,
	}, nil
}
