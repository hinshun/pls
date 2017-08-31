package ucpcmd

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/container"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/promise"
	"github.com/hinshun/pls/docker/hijack"
	"github.com/hinshun/pls/sshsession"
	"github.com/palantir/stacktrace"

	"gopkg.in/urfave/cli.v2"
)

const (
	authContainer       = "ucp-auth-api"
	sshPasswdCommand    = "docker exec -it ucp-auth-api enzi \"$(docker inspect --format '{{ index .Args 0 }}' ucp-auth-api)\" passwd -i"
	rethinkdbClientPort = "12383"
)

func Passwd(c *cli.Context) error {
	hostname := c.String("ssh")
	if hostname == "" {
		ctx := context.Background()
		cli, err := client.NewEnvClient()
		if err != nil {
			return stacktrace.Propagate(err, "failed to create docker client from env")
		}

		info, err := cli.Info(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "failed to get docker info")
		}

		nodeAddr := info.Swarm.NodeAddr
		execCfg := types.ExecConfig{
			Tty:          true,
			AttachStdin:  true,
			AttachStdout: true,
			Cmd:          []string{"enzi", fmt.Sprintf("--db-addr=%s:%s", nodeAddr, rethinkdbClientPort), "passwd", "-i"},
		}

		execResp, err := cli.ContainerExecCreate(ctx, authContainer, execCfg)
		if err != nil {
			return stacktrace.Propagate(err, "failed to create exec into container '%s'", authContainer)
		}

		hijackResp, err := cli.ContainerExecAttach(ctx, execResp.ID, execCfg)
		if err != nil {
			return stacktrace.Propagate(err, "failed to attach to exec into container '%s'", authContainer)
		}
		defer hijackResp.Close()

		dockerCLI := command.NewDockerCli(os.Stdin, os.Stdout, os.Stderr)
		err = dockerCLI.Initialize(flags.NewClientOptions())
		if err != nil {
			return stacktrace.Propagate(err, "failed to initialize docker cli")
		}

		errCh := promise.Go(func() error {
			streamer := hijack.New(dockerCLI, os.Stdin, os.Stdout, os.Stdout, hijackResp, execCfg.Tty, execCfg.DetachKeys)
			return streamer.Stream(ctx)
		})

		err = container.MonitorTtySize(ctx, dockerCLI, execResp.ID, true)
		if err != nil {
			return stacktrace.Propagate(err, "failed to monitor tty size")
		}

		err = <-errCh
		if err != nil {
			return stacktrace.Propagate(err, "received error during hijack")
		}

		return nil
	}

	keys := c.StringSlice("ssh-keypath")
	if len(keys) == 0 {
		keys = append(keys, os.Getenv("HOME")+"/.ssh/id_rsa")
	}

	sshSession, err := sshsession.New(hostname, keys)
	if err != nil {
		return stacktrace.Propagate(err, "failed to ssh into ucp manager")
	}
	defer sshSession.Close()

	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		return stacktrace.Propagate(err, "failed put terminal in raw mode")
	}

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		return stacktrace.Propagate(err, "failed get terminal width and height")
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := sshSession.RequestPty(os.Getenv("TERM"), termHeight, termWidth, modes); err != nil {
		return stacktrace.Propagate(err, "failed to request for pseudo terminal")
	}

	sshSession.Stdin = os.Stdin
	sshSession.Stdout = os.Stdout
	sshSession.Stderr = os.Stderr
	err = sshSession.Run(sshPasswdCommand)
	if err != nil {
		return stacktrace.Propagate(err, "failed to run passwd command over ssh")
	}

	err = terminal.Restore(fd, oldState)
	if err != nil {
		return stacktrace.Propagate(err, "failed to restore old terminal state")
	}

	return nil
}
