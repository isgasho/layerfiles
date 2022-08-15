package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/vm_protocol/vm_protocol_model"
	"io"
	"k8s.io/klog"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type ExposeWebsiteAddr struct {
	RequestId uint32
}

func (e *ExposeWebsiteAddr) Network() string {
	return "ws"
}

func (e *ExposeWebsiteAddr) String() string {
	return fmt.Sprintf("expose-website-addr (addr: %v)", e.RequestId)
}

type ExposeWebsiteConn struct {
	RequestId uint32
	Websocket *websocket.Conn
	Reader    io.Reader

	outBuf bytes.Buffer
}

func (e ExposeWebsiteConn) Read(b []byte) (n int, err error) {
	return e.Reader.Read(b)
}

func (e ExposeWebsiteConn) Write(b []byte) (n int, err error) {
	return e.outBuf.Write(b)
}

func (e ExposeWebsiteConn) Close() error {
	return e.Websocket.WriteMessage(websocket.BinaryMessage, e.outBuf.Bytes())
}

func (e ExposeWebsiteConn) LocalAddr() net.Addr {
	return &ExposeWebsiteAddr{RequestId: e.RequestId}
}

func (e ExposeWebsiteConn) RemoteAddr() net.Addr {
	return &ExposeWebsiteAddr{RequestId: e.RequestId}
}

func (e ExposeWebsiteConn) SetDeadline(t time.Time) error {
	return nil 	//do nothing
}

func (e ExposeWebsiteConn) SetReadDeadline(t time.Time) error {
	return nil 	//do nothing
}

func (e ExposeWebsiteConn) SetWriteDeadline(t time.Time) error {
	return nil 	//do nothing
}

type ExposeWebsiteListener struct {
	Websocket *websocket.Conn
}

func (e ExposeWebsiteListener) readExposeWebsiteMessage() (reqId uint32, data io.Reader, err error) {
	var reqIdBuf [4]byte
	msgType, reader, err := e.Websocket.NextReader()
	if msgType != websocket.BinaryMessage {
		return 0, nil, fmt.Errorf("got invalid message type %v from layerfile.com handler (from the layerfile side)", msgType)
	}
	if err != nil {
		return 0, nil, errors.Wrapf(err, "could not get next reader from websocket")
	}
	n, err := reader.Read(reqIdBuf[:])
	if err != nil {
		return 0, nil, errors.Wrapf(err, "could not read request id")
	}
	if n != 4 {
		return 0, nil, fmt.Errorf("invalid request, only %v/4 bytes of a request id", n)
	}

	reqId = binary.LittleEndian.Uint32(reqIdBuf[:])
	return reqId, reader, nil
}

func (e ExposeWebsiteListener) Accept() (net.Conn, error) {
	reqId, reader, err := e.readExposeWebsiteMessage()
	if err != nil {
		return nil, errors.Wrap(err, "could not read an expose website request from ws.layerfile.com")
	}
	if reqId == 0 {
		return nil, fmt.Errorf("unexpected control message from expose website handler")
	}

	conn := &ExposeWebsiteConn{RequestId: reqId, Reader: reader, Websocket: e.Websocket}

	var reqIdBuf [4]byte
	binary.LittleEndian.PutUint32(reqIdBuf[:], reqId)
	conn.outBuf.Write(reqIdBuf[:])
	return conn, nil
}

func (e ExposeWebsiteListener) Close() error {
	return e.Websocket.Close()
}

func (e ExposeWebsiteListener) Addr() net.Addr {
	return &ExposeWebsiteAddr{}
}

var _ net.Listener = (*ExposeWebsiteListener)(nil)

func (f *FuseFilewatcherClient) handleExposeWebsite(req *vm_protocol_model.ExposeWebsiteReq, key string) error {
	ws, _, err := websocket.DefaultDialer.Dial("wss://ws.layerfile.com", nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	var helloOut bytes.Buffer
	helloOut.Write([]byte{0,0,0,0}) //4 byte 'request id' is 0 for control messages
	json.NewEncoder(&helloOut).Encode(map[interface{}]interface{}{"key": key})
	err = ws.WriteMessage(websocket.BinaryMessage, helloOut.Bytes())
	if err != nil {
		return errors.Wrap(err, "could not write 'hello' message")
	}

	reverseProxy := httputil.ReverseProxy{}
	srv := http.Server{Handler: &reverseProxy}
	return srv.Serve(&ExposeWebsiteListener{Websocket: ws})
}

func (f *FuseFilewatcherClient) ExposeWebsite(ctx context.Context, req *vm_protocol_model.ExposeWebsiteReq) (*vm_protocol_model.ExposeWebsiteResp, error) {
	var keyBytes [16]byte
	_, err := rand.Read(keyBytes[:])
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	key := strings.Trim(base64.URLEncoding.EncodeToString(keyBytes[:]), "-_=")

	go func() {
		for {
			klog.Warning("error connecting for EXPOSE WEBSITE: ", f.handleExposeWebsite(req, key))
			time.Sleep(time.Second)
		}
	}()
	return &vm_protocol_model.ExposeWebsiteResp{Host: key+".layerfile.app"}, nil
}
