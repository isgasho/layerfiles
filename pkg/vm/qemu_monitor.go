package vm

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"os"
)

type QEMUMonitorHandler struct {
	conn net.Conn
}

func (mh *QEMUMonitorHandler) Connect(port int) error {
	var err error
	mh.conn, err = net.Dial("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(mh.conn)
		firstLine := true
		for scanner.Scan() {
			if firstLine {
				firstLine = false
				continue
			}
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	return nil
}

func (mh *QEMUMonitorHandler) SendCommand(command string) error {
	_, err := mh.conn.Write([]byte(command + "\n"))
	return errors.Wrapf(err, "could not send %+v command", command)
}
