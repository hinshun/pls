package main

import (
	"os"

	"github.com/hinshun/pls/command/dindcmd"
	"github.com/hinshun/pls/command/mitmcmd"
	"github.com/hinshun/pls/command/rethinkdbcmd"
	"github.com/hinshun/pls/command/ucpcmd"
	"github.com/hinshun/pls/docker/dind"

	"gopkg.in/urfave/cli.v2"
)

func main() {
	app := &cli.App{
		Name:  "pls",
		Usage: "pretty please",
		Commands: []*cli.Command{
			{
				Name:  "ucp",
				Usage: "Manage UCP",
				Subcommands: []*cli.Command{
					{
						Name:  "passwd",
						Usage: "Change the admin username and password",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "ssh",
								Usage: "The SSH hostname of a UCP manager to perform passwd over SSH",
							},
							&cli.StringSliceFlag{
								Name:  "ssh-keypath",
								Usage: "Specify a path to a private key to be used to authenticate the SSH connection",
							},
						},
						Action: WrapAction(ucpcmd.Passwd),
					},
				},
			},
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
								Name:  "image",
								Usage: "The image to run Docker in Docker",
								Value: dind.DindImageName,
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
						Action: WrapAction(dindcmd.CreateDind),
					},
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "List dind containers",
						Action:  WrapAction(dindcmd.ListDinds),
					},
					{
						Name:   "prune",
						Usage:  "Remove all dind containers",
						Action: WrapAction(dindcmd.PruneDinds),
					},
				},
			},
			{
				Name:  "mitm",
				Usage: "Manage mitmproxy (Man-in-the-Middle) containers",
				Subcommands: []*cli.Command{
					{
						Name:  "create",
						Usage: "Create a new mitmproxy container",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "name",
								Usage: "Assign a name to the container",
							},
						},
						Action: WrapAction(mitmcmd.CreateMITMProxy),
					},
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "List mitmproxy containers",
						Action:  WrapAction(mitmcmd.ListMITMProxies),
					},
					{
						Name:   "prune",
						Usage:  "Remove all mitmproxy containers",
						Action: WrapAction(mitmcmd.PruneMITMProxies),
					},
				},
			},
			{
				Name:  "rethinkdb",
				Usage: "Manage RethinkDB",
				Subcommands: []*cli.Command{
					{
						Name:   "repl",
						Usage:  "Creates a ReQL REPL to a RethinkDB",
						Action: WrapAction(rethinkdbcmd.CreateRethinkdbREPL),
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
			return err
		}
		return nil
	}
}

// type cliError struct {
// 	err error
// }

// func (c cliError) Error() string {
// 	return c.err.Error()
//  return stacktrace.RootCause(c.err).Error()
// }
