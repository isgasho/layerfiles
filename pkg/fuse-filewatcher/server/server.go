package server

import (
	"context"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/fuse-filewatcher/filewatcher_model"
	"google.golang.org/grpc"
	"io"
	"k8s.io/klog"
	"net"
	"os"
	"time"
)

type FuseFilewatcherServer struct {
	filewatcher_model.UnimplementedFuseFSServer
	RPCListenAddr string
	MetaHost      string

	OnAccess func(path string)
	OnRead   func(path string)

	rpcListener net.Listener
	grpcServer  *grpc.Server
	metaConn    *grpc.ClientConn
	metaClient  filewatcher_model.FuseFilewatcherClientClient
	started     bool
}

func NewFuseFilewatcherServer() *FuseFilewatcherServer {
	return &FuseFilewatcherServer{
		OnRead:   func(string) {},
		OnAccess: func(string) {},
	}
}

//AllowReads : stop blocking reads (e.g., when repo is done cloning)
func (s *FuseFilewatcherServer) AllowReads() error {
	_, err := s.WaitForConn().AllowReads(context.Background(), &filewatcher_model.AllowReadsReq{})
	return err
}

func (s *FuseFilewatcherServer) NotifyAccess(ctx context.Context, req *filewatcher_model.NotifyAccessReq) (*filewatcher_model.NotifyAccessResp, error) {
	s.OnAccess(req.Path)
	return &filewatcher_model.NotifyAccessResp{}, nil
}

func (s *FuseFilewatcherServer) NotifyRead(ctx context.Context, req *filewatcher_model.NotifyReadReq) (*filewatcher_model.NotifyReadResp, error) {
	s.OnRead(req.Path)
	return &filewatcher_model.NotifyReadResp{}, nil
}

func (s *FuseFilewatcherServer) ReadFile(req *filewatcher_model.ReadFileReq, srv filewatcher_model.FuseFS_ReadFileServer) error {
	var resp filewatcher_model.ReadFileResp

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

func (s *FuseFilewatcherServer) ReadDir(ctx context.Context, req *filewatcher_model.ReadDirReq) (*filewatcher_model.ReadDirResp, error) {
	resp := &filewatcher_model.ReadDirResp{}

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
	resp.Entries = make([]*filewatcher_model.Dirent, len(files))
	for i, f := range files {
		resp.Entries[i] = &filewatcher_model.Dirent{
			Name: f.Name(),
			IsDir: f.IsDir(),
		}
	}
	return resp, nil
}

func (s *FuseFilewatcherServer) Started() bool {
	return s.started
}

func (s *FuseFilewatcherServer) WaitForConn() filewatcher_model.FuseFilewatcherClientClient {
	for i := 0; i < 100; i += 1 {
		if s.metaClient != nil {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}
	return s.metaClient
}

func (s *FuseFilewatcherServer) Run() error {
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
				s.metaClient = filewatcher_model.NewFuseFilewatcherClientClient(s.metaConn)
				break
			}
			klog.Info("Waiting for client to be up: ", err)
			time.Sleep(time.Millisecond * 100)
		}
	}()

	s.grpcServer = grpc.NewServer()
	filewatcher_model.RegisterFuseFSServer(s.grpcServer, s)
	klog.Info("Done setting up RPC server (server)")
	s.started = true
	return s.grpcServer.Serve(s.rpcListener)
}

func (s *FuseFilewatcherServer) Stop() {
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
