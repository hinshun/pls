package main

import (
	"os"

	"github.com/hinshun/pls/command"

	"gopkg.in/urfave/cli.v2"
)

func main() {
	app := &cli.App{
		Name:  "pls",
		Usage: "pretty please",
		Commands: []*cli.Command{
			{
				Name: "dind",
				Subcommands: []*cli.Command{
					{
						Name:  "create",
						Usage: "Create a new dind container",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "mitm",
								Usage: "Man-in-the-Middle proxy to intercept outgoing dockerd traffic",
							},
						},
						Action: WrapAction(command.CreateDind),
					},
					{
						Name:   "prune",
						Usage:  "Remove all dind containers",
						Action: WrapAction(command.PruneDinds),
					},
				},
			},
			{
				Name: "mitm",
				Subcommands: []*cli.Command{
					{
						Name:   "create",
						Usage:  "Create a new mitmproxy container",
						Action: WrapAction(command.CreateMITMProxy),
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
