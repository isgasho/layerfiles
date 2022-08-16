// +build linux

package qemu

import (
	"bytes"
	_ "embed"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed qemu-system-x86_64
var QemuBinary []byte

func QemuCommand(Args ...string) (*exec.Cmd, error) {
	executablePath := filepath.Join(os.TempDir(), "layerfile-qemu-system-x86_64")
	f, err := os.OpenFile(executablePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0700)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open the file at %v", executablePath)
	}

	_, err = io.Copy(f, bytes.NewReader(QemuBinary))
	if err != nil {
		return nil, errors.Wrapf(err, "could write binary file at %v", executablePath)
	}

	err = f.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "could flush binary file at %v", executablePath)
	}

	return exec.Command(executablePath, Args...), nil
}