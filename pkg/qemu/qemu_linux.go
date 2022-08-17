// +build linux

package qemu

import (
	"bytes"
	_ "embed"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/environment"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed qemu-system-x86_64
var QemuBinary []byte

//go:embed qboot.rom
var QBootBinary []byte

func ensureQbootExists(dest string) error {
	f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0700)
	if err != nil {
		return errors.Wrapf(err, "could not open the qboot.bin file at %v", dest)
	}

	_, err = io.Copy(f, bytes.NewReader(QBootBinary))
	if err != nil {
		return errors.Wrapf(err, "could write qboot.bin file at %v", dest)
	}

	err = f.Close()
	if err != nil {
		return errors.Wrapf(err, "could flush qboot.bin file at %v", dest)
	}
	return nil
}

func QemuCommand(Args ...string) (*exec.Cmd, error) {
	binDir, err := environment.GetAndCreateBinDirectory()
	if err != nil {
		return nil, err
	}

	executablePath := filepath.Join(binDir, "layerfile-qemu-system-x86_64")
	f, err := os.OpenFile(executablePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0700)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open the file at %v", executablePath)
	}

	qbootPath := filepath.Join(binDir, "qboot.rom") //TODO doesn't really belong in bin dir?
	err = ensureQbootExists(qbootPath)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(f, bytes.NewReader(QemuBinary))
	if err != nil {
		return nil, errors.Wrapf(err, "could write binary file at %v", executablePath)
	}

	err = f.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "could flush binary file at %v", executablePath)
	}

	biosSet := false
	for _, arg := range Args {
		if arg == "-bios" {
			biosSet = true
		}
	}
	if !biosSet {
		Args = append(Args, "-bios", qbootPath)
	}

	return exec.Command(executablePath, Args...), nil
}