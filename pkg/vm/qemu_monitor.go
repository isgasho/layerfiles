package vm

import (
	"fmt"
	"github.com/pkg/errors"
	"net"
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

	return nil
}

func (mh *QEMUMonitorHandler) SystemReset() error {
	_, err := mh.conn.Write([]byte("system_reset\n"))
	return errors.Wrapf(err, "could not send system reset command")
}