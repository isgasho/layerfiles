package vm

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/grepv"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type QEMUCommandHandler struct {
	Stdout io.ReadCloser
	Stdin io.WriteCloser

	outBuf bytes.Buffer
	cleanWriter io.Writer
}

func (handler *QEMUCommandHandler) WaitForRegex(regex *regexp.Regexp, writer io.Writer) ([][]byte, error) {
	var buf [65536]byte
	for {
		n, err := handler.Stdout.Read(buf[:])
		handler.outBuf.Write(buf[:n])
		if writer != nil {
			_, err := writer.Write(buf[:n])
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

type CommandStatusCodeError struct {
	StatusCode int
	Output []byte
}

func (err *CommandStatusCodeError) Error() string {
	if len(err.Output) + len(err.Output) == 0 {
		return fmt.Sprintf("Command exited with status code %v", err.StatusCode)
	}
	firstline := strings.SplitN(strings.TrimSpace(string(err.Output)), "\n", 2)[0]
	if len(firstline) > 131 {
		firstline = firstline[:128] + fmt.Sprintf("... (%v more bytes)", len(firstline)-128)
	}
	return fmt.Sprintf("Status code %v: %v", err.StatusCode, firstline)
}

var commandDoneStatusRegex = regexp.MustCompile("RGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu\\s+(\\d+)")

func (handler *QEMUCommandHandler) RunCommand(cmd string) (out string, err error) {
	cmd += "; echo -e '\\n\\nRGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu' $? 'nRGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu' "

	var cmdOutput bytes.Buffer

	handler.Stdin.Write([]byte(cmd + "\n"))
	match, err := handler.WaitForRegex(commandDoneStatusRegex, &cmdOutput)
	if err != nil {
		return "", err
	}

	statusCode, err := strconv.ParseInt(strings.TrimSpace(string(match[1])), 10, 32)
	if err != nil {
		return "", errors.Wrap(err, "non-numeric exit status")
	}

	if statusCode == 0 {
		return cmdOutput.String(), nil
	}

	return cmdOutput.String(), &CommandStatusCodeError{StatusCode: int(statusCode), Output: cmdOutput.Bytes()}
}

func (handler *QEMUCommandHandler) RunCommandStreamOutput(cmd string) error {
	cmd = strings.TrimSpace(cmd)
	if !strings.HasSuffix(cmd, ";") {
		cmd = cmd + ";" //hack
	}
	cmd += " echo -e '\\n\\nRGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu' $? 'nRGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu' "

	if handler.cleanWriter == nil {
		handler.cleanWriter = grepv.New([]byte("\r\n\r\nRGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu"), []byte("nRGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu"), os.Stdout)
	}

	handler.Stdin.Write([]byte(cmd + "\n"))

	match, err := handler.WaitForRegex(commandDoneStatusRegex, handler.cleanWriter)
	if err != nil {
		return err
	}

	statusCode, err := strconv.ParseInt(strings.TrimSpace(string(match[1])), 10, 32)
	if err != nil {
		return errors.Wrap(err, "non-numeric exit status")
	}

	if statusCode == 0 {
		return nil
	}
	return &CommandStatusCodeError{StatusCode: int(statusCode), Output: []byte("(no output - streaming)")}

}