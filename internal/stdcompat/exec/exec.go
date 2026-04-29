// SPDX-License-Identifier: EUPL-1.2

package exec

import (
	"context"
	"syscall"
	"time"

	core "dappco.re/go"
)

type Cmd struct {
	ctx    context.Context
	name   string
	args   []string
	Stdout core.Writer
	Stderr core.Writer
}

func CommandContext(ctx context.Context, name string, args ...string) *Cmd {
	return &Cmd{ctx: ctx, name: name, args: args, Stdout: core.Stdout(), Stderr: core.Stderr()}
}

func (c *Cmd) Run() error {
	if c == nil {
		return core.NewError("nil command")
	}
	argv := append([]string{c.name}, c.args...)
	command := c.name
	if found := (core.App{}).Find(c.name, c.name); found.OK {
		if app, ok := found.Value.(*core.App); ok && app.Path != "" {
			command = app.Path
		}
	}
	pid, err := syscall.ForkExec(command, argv, &syscall.ProcAttr{
		Env:   core.Environ(),
		Files: []uintptr{0, 1, 2},
	})
	if err != nil {
		return err
	}
	for {
		var status syscall.WaitStatus
		done, waitErr := syscall.Wait4(pid, &status, syscall.WNOHANG, nil)
		if waitErr != nil {
			return waitErr
		}
		if done == pid {
			if status.ExitStatus() == 0 {
				return nil
			}
			return core.Errorf("exit status %d", status.ExitStatus())
		}
		select {
		case <-c.ctx.Done():
			if killErr := syscall.Kill(pid, syscall.SIGKILL); killErr != nil {
				return killErr
			}
			return c.ctx.Err()
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
