package main

import (
	"os"

	"github.com/hinshun/pls/command"
	"github.com/palantir/stacktrace"

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
						Name:   "create",
						Usage:  "Create a new dind container",
						Action: WrapAction(command.CreateDind),
					},
					{
						Name:   "prune",
						Usage:  "Remove all dind containers",
						Action: WrapAction(command.PruneDinds),
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
	return stacktrace.RootCause(c.err).Error()
}
