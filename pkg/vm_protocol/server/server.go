package server

import (
	"context"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/vm_protocol/vm_protocol_model"
	"google.golang.org/grpc"
	"io"
	"k8s.io/klog"
	"net"
	"os"
	"time"
)

type VMContactServer struct {
	vm_protocol_model.UnimplementedVMProtocolServerServer
	RPCListenAddr string
	MetaHost      string

	OnAccess func(path string)
	OnRead   func(path string)

	rpcListener net.Listener
	grpcServer  *grpc.Server
	metaConn    *grpc.ClientConn
	metaClient  vm_protocol_model.VMProtocolClientClient
	started     bool
}

func NewFuseFilewatcherServer() *VMContactServer {
	return &VMContactServer{
		OnRead:   func(string) {},
		OnAccess: func(string) {},
	}
}

//AllowReads : stop blocking reads (e.g., when repo is done cloning)
func (s *VMContactServer) AllowReads() error {
	_, err := s.WaitForConn().AllowReads(context.Background(), &vm_protocol_model.AllowReadsReq{})
	return err
}

func (s *VMContactServer) NotifyAccess(ctx context.Context, req *vm_protocol_model.NotifyAccessReq) (*vm_protocol_model.NotifyAccessResp, error) {
	s.OnAccess(req.Path)
	return &vm_protocol_model.NotifyAccessResp{}, nil
}

func (s *VMContactServer) NotifyRead(ctx context.Context, req *vm_protocol_model.NotifyReadReq) (*vm_protocol_model.NotifyReadResp, error) {
	s.OnRead(req.Path)
	return &vm_protocol_model.NotifyReadResp{}, nil
}

func (s *VMContactServer) ReadFile(req *vm_protocol_model.ReadFileReq, srv vm_protocol_model.VMProtocolServer_ReadFileServer) error {
	var resp vm_protocol_model.ReadFileResp

	processError := func(err error) {
		if os.IsNotExist(err) {
			resp.ErrorIsNotExist = true
		}
		if errors.Cause(err) != io.EOF {
			resp.Error = err.Error()
			resp.Data = resp.Data[:0]
		}

		srv.Send(&resp)
	}

	file, err := os.Open(req.Path)
	if err != nil {
		processError(err)
		return nil
	}
	defer file.Close()

	var buf [1024 * 64]byte
	for {
		n, err := file.Read(buf[:])
		resp.Data = buf[:n]
		if err != nil {
			processError(err)
			return nil
		}
		srv.Send(&resp)
	}
}

func (s *VMContactServer) ReadDir(ctx context.Context, req *vm_protocol_model.ReadDirReq) (*vm_protocol_model.ReadDirResp, error) {
	resp := &vm_protocol_model.ReadDirResp{}

	stat, err := os.Stat(req.Path)
	if err != nil {
		if os.IsNotExist(err) {
			resp.ErrorIsNotExist = true
		}
		resp.Error = err.Error()
		return resp, nil
	}
	resp.Mode = uint64(stat.Mode())
	if !stat.IsDir() {
		resp.ErrorIsNotDir = true
		resp.Error = "not a directory"
		return resp, nil
	}

	files, err := os.ReadDir(req.Path)
	if err != nil {
		if os.IsNotExist(err) {
			resp.ErrorIsNotExist = true
		}
		resp.Error = err.Error()
		return resp, nil
	}
	resp.Entries = make([]*vm_protocol_model.Dirent, len(files))
	for i, f := range files {
		resp.Entries[i] = &vm_protocol_model.Dirent{
			Name: f.Name(),
			IsDir: f.IsDir(),
		}
	}
	return resp, nil
}

func (s *VMContactServer) Started() bool {
	return s.started
}

func (s *VMContactServer) WaitForConn() vm_protocol_model.VMProtocolClientClient {
	for i := 0; i < 100; i += 1 {
		if s.metaClient != nil {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}
	return s.metaClient
}

func (s *VMContactServer) Run() error {
	defer s.Stop()

	var err error
	s.rpcListener, err = net.Listen("tcp", s.RPCListenAddr)
	if err != nil {
		return err
	}

	go func() {
		for {
			s.metaConn, err = grpc.DialContext(context.Background(), s.MetaHost, grpc.WithInsecure(), grpc.WithBlock())
			if err == nil {
				s.metaClient = vm_protocol_model.NewVMProtocolClientClient(s.metaConn)
				break
			}
			klog.Info("Waiting for client to be up: ", err)
			time.Sleep(time.Millisecond * 100)
		}
	}()

	s.grpcServer = grpc.NewServer()
	vm_protocol_model.RegisterVMProtocolServerServer(s.grpcServer, s)
	klog.Info("Done setting up RPC server (server)")
	s.started = true
	return s.grpcServer.Serve(s.rpcListener)
}

func (s *VMContactServer) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.rpcListener != nil {
		s.rpcListener.Close()
	}
	if s.metaConn != nil {
		s.metaConn.Close()
	}
}
