package client

import (
	"context"
	"fmt"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/fuse-filewatcher/client/filesystems"
	"github.com/webappio/layerfiles/pkg/fuse-filewatcher/filewatcher_model"
	"google.golang.org/grpc"
	"io"
	"k8s.io/klog"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

type FuseFilewatcherClient struct {
	filewatcher_model.UnimplementedFuseFilewatcherClientServer

	MetaListenAddr string //messages from vm-worker -> vm
	RPCHost        string //messages from vm -> vm-worker

	Source string //usually /var/lib/webappio/copied-repo-files
	Dest   string //usually /var/lib/webappio/repo-files

	accessQueue chan string
	readQueue   chan string

	copyRequestsPending uint32

	lastPing time.Time

	metaListener net.Listener
	metaServer   *grpc.Server
	RPCConn      *grpc.ClientConn

	stopped   bool
	started   bool
	reconnect bool

	client     filewatcher_model.FuseFSClient
	loopbackFS *filesystems.LoopbackFileSystem
}

func (f *FuseFilewatcherClient) Reconnect(ctx context.Context, req *filewatcher_model.ReconnectReq) (*filewatcher_model.ReconnectResp, error) {
	if f.RPCConn != nil {
		f.RPCConn.Close()
	}
	var err error
	f.RPCConn, err = grpc.DialContext(ctx, f.RPCHost, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		klog.Warning(errors.Wrapf(err, "could not connect to rpc host %v", f.RPCHost))
		return nil, err
	}

	return &filewatcher_model.ReconnectResp{}, nil
}

func (f *FuseFilewatcherClient) AllowReads(context.Context, *filewatcher_model.AllowReadsReq) (*filewatcher_model.AllowReadsResp, error) {
	klog.V(3).Info("Got an AllowReads request from the vm-worker")
	f.loopbackFS.AllowReads()
	return &filewatcher_model.AllowReadsResp{}, nil
}

func (f *FuseFilewatcherClient) Sync(ctx context.Context, req *filewatcher_model.SyncReq) (*filewatcher_model.SyncResp, error) {
	for (len(f.accessQueue)+len(f.readQueue) > 0 || f.reconnect || f.copyRequestsPending > 0) && ctx.Err() == nil {
		time.Sleep(time.Millisecond * 50)
	}
	return &filewatcher_model.SyncResp{}, ctx.Err()
}

func (f *FuseFilewatcherClient) Copy(ctx context.Context, req *filewatcher_model.CopyReq) (*filewatcher_model.CopyResp, error) {
	atomic.AddUint32(&f.copyRequestsPending, 1)
	defer func() {
		atomic.AddUint32(&f.copyRequestsPending, ^uint32(0)) //this is how docs say to decrement lol?
	}()

	copyFile := func(src, dest string) error {
		resp, err := f.client.ReadFile(ctx, &filewatcher_model.ReadFileReq{Path: src})
		if err != nil {
			return errors.Wrapf(err, "could not execute ReadFile RPC for %v", src)
		}

		var file *os.File

		var msg filewatcher_model.ReadFileResp
		eof := false
		for !eof {
			err := resp.RecvMsg(&msg)
			if err != nil && err != io.EOF {
				return errors.Wrapf(err, "could not execute RecvMsg for ReadFile on %v", src)
			}
			eof = err == io.EOF
			if file == nil {
				file, err = os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(msg.GetMode()))
				if err != nil {
					return errors.Wrapf(err, "could not open file at %v", dest)
				}
				defer func() {
					file.Close()
				}()
			}
			_, err = file.Write(msg.Data)
			if err != nil {
				return errors.Wrapf(err, "could not write to file at %v", dest)
			}
		}
		return nil
	}

	var copyDirOrFile func(src, dest string) error
	copyDirOrFile = func(src, dest string) error {
		resp, err := f.client.ReadDir(ctx, &filewatcher_model.ReadDirReq{Path: src})
		if err != nil {
			return errors.Wrapf(err, "could not execute ReadDir RPC for %v", src)
		}
		if resp.Error == "" {
			err := os.Mkdir(dest, os.FileMode(resp.Mode))
			if err != nil {
				return errors.Wrapf(err, "could not create directory at %v", dest)
			}
			for _, dir := range resp.Entries {
				if dir.IsDir {
					err = copyDirOrFile(filepath.Join(src, dir.Name), filepath.Join(dest, dir.Name))
				} else {
					err = copyFile(filepath.Join(src, dir.Name), filepath.Join(dest, dir.Name))
				}
				if err != nil {
					return err
				}
			}
		}
		if resp.ErrorIsNotDir {
			return copyFile(src, dest)
		}
		if resp.ErrorIsNotExist {
			return os.ErrNotExist
		}
		return fmt.Errorf("error copying directory: %v", resp.Error)
	}

	src := req.HostSource
	dest := filepath.Join("/var/lib/layerfiles/copied-repo-files", filepath.Base(req.HostSource))
	err := os.MkdirAll(dest, 0o755)
	if err != nil {
		return &filewatcher_model.CopyResp{Error: err.Error()}, nil
	}
	err = copyDirOrFile(src, dest)
	if err != nil {
		return &filewatcher_model.CopyResp{Error: err.Error()}, nil
	}
	return &filewatcher_model.CopyResp{}, nil
}

func (f *FuseFilewatcherClient) connectAndRun() {
	var err error
	klog.Info("Client connecting...")

	f.RPCConn, err = grpc.Dial(f.RPCHost, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		klog.Warning(errors.Wrapf(err, "could not connect to rpc host %v", f.RPCHost))
		return
	}
	defer f.RPCConn.Close()
	klog.Info("Connected server->client conn.")

	f.client = filewatcher_model.NewFuseFSClient(f.RPCConn)

	f.metaListener, err = net.Listen("tcp", f.MetaListenAddr)
	if err != nil {
		klog.Error(err)
		return
	}
	defer f.metaListener.Close()
	klog.Info("Connected client->server conn.")

	f.metaServer = grpc.NewServer()
	filewatcher_model.RegisterFuseFilewatcherClientServer(f.metaServer, f)
	defer f.metaServer.Stop()

	rpcEnded := false
	go func() {
		err := f.metaServer.Serve(f.metaListener)
		if err != nil {
			klog.Warning(err)
		}
		rpcEnded = true
	}()

	f.started = true
	klog.Info("Done setting up filewatcher (client)")
	f.reconnect = false
	for !f.stopped && !rpcEnded && !f.reconnect {
		time.Sleep(time.Millisecond * 350)
	}
	klog.Info("Client disconnected.")
}

func (f *FuseFilewatcherClient) Started() bool {
	return f.started
}

func (f *FuseFilewatcherClient) createMount() error {
	err := os.MkdirAll(f.Dest, 0755)
	if err != nil {
		return errors.Wrapf(err, "could not create files at %v", f.Dest)
	}

	f.loopbackFS, err = filesystems.NewLoopbackFileSystem(f.Source)
	if err != nil {
		return errors.Wrapf(err, "could not create loopback FS for existing file")
	}
	f.loopbackFS.OnRead = func(path string) {
		f.readQueue <- path
	}
	f.loopbackFS.OnAccess = func(path string) {
		f.accessQueue <- path
	}

	oneDay := time.Hour * 24
	opts := &fs.Options{
		MountOptions: fuse.MountOptions{
			Options:      []string{"max_read=131072", "default_permissions", "nonempty"},
			MaxReadAhead: 128 << 10, //128kb
			FsName:       "layer-dir-watcher",
			Name:         "layerdir",
			AllowOther:   true,
			EnableLocks:  false,
			Debug:        bool(klog.V(6)),
		},
		AttrTimeout:     &oneDay,
		NegativeTimeout: &oneDay,
		EntryTimeout:    &oneDay,
	}

	_, err = fs.Mount(f.Dest, f.loopbackFS.Root(), opts)
	if err != nil {
		return errors.Wrapf(err, "could not fs.Mount(%v)", f.Dest)
	}

	klog.Info("Successfully created mount at ", f.Dest)
	return nil
}

func (f *FuseFilewatcherClient) Run() {
	f.accessQueue = make(chan string, 1024)
	f.readQueue = make(chan string, 1024)
	go func() {
		for !f.stopped {
			select {
			case path := <-f.accessQueue:
				for {
					_, err := f.client.NotifyAccess(context.Background(), &filewatcher_model.NotifyAccessReq{Path: path})
					if err == nil {
						break
					}
					klog.Warning(err)
					f.reconnect = true
					time.Sleep(time.Millisecond * 50)
				}
			case path := <-f.readQueue:
				for {
					_, err := f.client.NotifyRead(context.Background(), &filewatcher_model.NotifyReadReq{Path: path})
					if err == nil {
						break
					}
					klog.Warning(err)
					f.reconnect = true
					time.Sleep(time.Millisecond * 50)
				}
			}
		}
		klog.Info("Client stopped.")
	}()
	klog.Info("Creating mount.")
	err := f.createMount()
	if err != nil {
		klog.Error("Could not create mount: ", err)
		panic(err)
	}
	klog.Info("Directory mounted, connecting to server.")
	for !f.stopped {
		f.connectAndRun()
		time.Sleep(time.Millisecond * 100)
	}
}

func (f *FuseFilewatcherClient) Stop() {
	klog.Info("FuseFilewatcherClient stopping.")
	f.stopped = true
	if f.metaListener != nil {
		f.metaListener.Close()
	}
	if f.RPCConn != nil {
		f.RPCConn.Close()
	}
}
