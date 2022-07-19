package ezdc

import (
	"context"
	"os"
	"os/exec"
)

// compose command
type cc struct {
	project string
	file    string
}

func (c cc) cmd(ctx context.Context, args ...string) *exec.Cmd {
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

func (c cc) down(ctx context.Context) *exec.Cmd {

	return c.cmd(ctx, "down", "-v", "--remove-orphans", "--rmi", "local", "--timeout", "0")
}
func (c cc) pull(ctx context.Context, svcs ...string) *exec.Cmd {
	if len(svcs) == 0 {
		return nil
	}
	return c.cmd(ctx, append([]string{"pull"}, svcs...)...)
}
func (c cc) build(ctx context.Context) *exec.Cmd {
	return c.cmd(ctx, "build")
}

func (c cc) stop(ctx context.Context) *exec.Cmd {
	return c.cmd(ctx, "stop", "--timeout", "0")
}

func (c cc) up(ctx context.Context) *exec.Cmd {
	return c.cmd(ctx, "up")
}
