package util

import (
	"bytes"
	"github.com/pkg/errors"
	"os/exec"
	"strings"
	"time"
)

func RunCommandWithTimeout(command *exec.Cmd, timeout time.Duration) error {
	var out bytes.Buffer
	if command.Stdout == nil {
		command.Stdout = &out
	}
	if command.Stderr == nil {
		command.Stderr = &out
	}
	commandDone := make(chan error)
	commandCancelled := time.NewTimer(timeout)
	defer commandCancelled.Stop()

	err := command.Start()
	if err != nil {
		return errors.Wrap(err, "could not start "+command.Args[0])
	}

	go func() {
		commandDone <- command.Wait()
	}()

	select {
	case err := <-commandDone:
		return errors.Wrap(err, strings.TrimSpace(out.String()))
	case <-commandCancelled.C:
		command.Process.Kill()
		return errors.New("timed out after " + timeout.String() + " running " + command.Args[0])
	}
}
