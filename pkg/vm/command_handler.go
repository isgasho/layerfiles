package vm

import (
	"bytes"
	"github.com/pkg/errors"
	"io"
	"regexp"
)

type QEMUCommandHandler struct {
	Stdout io.ReadCloser
	Stdin io.WriteCloser

	outBuf bytes.Buffer
	cleanWriter io.Writer
}

func (handler *QEMUCommandHandler) WaitForRegex(regex *regexp.Regexp, printOutput bool) ([][]byte, error) {
	var buf [65536]byte
	for {
		n, err := handler.Stdout.Read(buf[:])
		handler.outBuf.Write(buf[:n])
		if printOutput {
			_, err := handler.cleanWriter.Write(buf[:n])
			if err != nil {
				return nil, err
			}
		}

		if err == io.EOF {
			return nil, errors.New("The VM running this test crashed.")
		} else if err != nil {
			return nil, errors.Wrap(err, "could not read output from last command run")
		}

		loc := regex.FindIndex(handler.outBuf.Bytes())
		if loc != nil {
			submatches := regex.FindSubmatch(handler.outBuf.Bytes()[loc[0]:loc[1]])
			handler.outBuf.Next(loc[1])
			return submatches, nil
		}
	}
}