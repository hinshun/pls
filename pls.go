package main

import (
	"os"

	"github.com/hinshun/pls/command"
	"github.com/hinshun/pls/docker/dind"

	"gopkg.in/urfave/cli.v2"
)

func main() {
	app := &cli.App{
		Name:  "pls",
		Usage: "pretty please",
		Commands: []*cli.Command{
			{
				Name:  "dind",
				Usage: "Manage Docker in Docker containers",
				Subcommands: []*cli.Command{
					{
						Name:  "create",
						Usage: "Create a new dind container",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "name",
								Usage: "Assign a name to the container",
							},
							&cli.StringFlag{
								Name:  "mitm",
								Usage: "Proxy outgoing dockerd traffic to mitmproxy",
							},
							&cli.StringFlag{
								Name:  "registry",
								Usage: "The server address of a Docker Registry to trust",
								Value: dind.DefaultRegistryServerAddress,
							},
							&cli.StringFlag{
								Name:  "username",
								Usage: "The username to authenticate against the Docker Registry",
							},
							&cli.StringFlag{
								Name:  "password",
								Usage: "The password to authenticate against the Docker Registry",
							},
						},
						Action: WrapAction(command.CreateDind),
					},
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "List dind containers",
						Action:  WrapAction(command.ListDinds),
					},
					{
						Name:   "prune",
						Usage:  "Remove all dind containers",
						Action: WrapAction(command.PruneDinds),
					},
				},
			},
			{
				Name:  "mitm",
				Usage: "Manage mitmproxy (Man-in-the-Middle) containers",
				Subcommands: []*cli.Command{
					{
						Name: "create",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "name",
								Usage: "Assign a name to the container",
							},
						},
						Usage:  "Create a new mitmproxy container",
						Action: WrapAction(command.CreateMITMProxy),
					},
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "List mitmproxy containers",
						Action:  WrapAction(command.ListMITMProxies),
					},
					{
						Name:   "prune",
						Usage:  "Remove all mitmproxy containers",
						Action: WrapAction(command.PruneMITMProxies),
					},
				},
			},
		},
	}

	app.Run(os.Args)
}

func WrapAction(action cli.ActionFunc) cli.ActionFunc {
	return func(c *cli.Context) error {
		err := action(c)
		if err != nil {
			return cliError{err}
		}
		return nil
	}
}

type cliError struct {
	err error
}

func (c cliError) Error() string {
	return c.err.Error()
	// return stacktrace.RootCause(c.err).Error()
}
