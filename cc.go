package ezdc

import (
	"context"
	"os"
	"os/exec"
)

type ComposeFace interface {
	Up(ctx context.Context) *exec.Cmd
	Down(ctx context.Context) *exec.Cmd
	Stop(ctx context.Context) *exec.Cmd
	Pull(ctx context.Context, svc ...string) *exec.Cmd
	Build(ctx context.Context) *exec.Cmd
}

type DefaultComposeCmd struct {
	project string
	file    string
}

func (c DefaultComposeCmd) cmd(ctx context.Context, args ...string) *exec.Cmd {
	file := c.file
	if file == "" {
		file = "./docker-compose.yml"
	}
	cmd := exec.CommandContext(ctx, "docker",
		append([]string{"compose", "-p", c.project, "-f", file}, args...)...,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func (c DefaultComposeCmd) Down(ctx context.Context) *exec.Cmd {

	return c.cmd(ctx, "down", "-v", "--remove-orphans", "--rmi", "local", "--timeout", "0")
}
func (c DefaultComposeCmd) Pull(ctx context.Context, svcs ...string) *exec.Cmd {
	if len(svcs) == 0 {
		return nil
	}
	return c.cmd(ctx, append([]string{"pull"}, svcs...)...)
}
func (c DefaultComposeCmd) Build(ctx context.Context) *exec.Cmd {
	return c.cmd(ctx, "build")
}

func (c DefaultComposeCmd) Stop(ctx context.Context) *exec.Cmd {
	return c.cmd(ctx, "stop", "--timeout", "0")
}

func (c DefaultComposeCmd) Up(ctx context.Context) *exec.Cmd {
	return c.cmd(ctx, "up")
}
