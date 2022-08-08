package file_share_server

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/fuse-filewatcher/filewatcher_model"
	fusefilewatcherserver "github.com/webappio/layerfiles/pkg/fuse-filewatcher/server"
	"github.com/webappio/layerfiles/pkg/vm"
	"log"
)

type FileShareServer struct {
	running bool

	fuseServer *fusefilewatcherserver.FuseFilewatcherServer
}

func (server *FileShareServer) ensureFuseClientRunning(vmI *vm.QemuVM) error {
	cmdHandler := vmI.GetCommandHandler()
	_, err := cmdHandler.RunCommand("pgrep -f fuse-filewatcher-v2-linux-amd64")
	if err == nil {
		return nil //it's already running
	} else if err, ok := err.(*vm.CommandStatusCodeError); ok && err.StatusCode == 1 {
		//it's not running
	} else {
		return err //some other error
	}

	_, err = cmdHandler.RunCommand("cd / && ulimit -n 65535 && " +
		fmt.Sprintf("bash -c 'nice -n -15 nohup /usr/local/bin/fuse-filewatcher-v2-linux-amd64 :30812 %v:30811 /var/lib/layerfiles/copied-repo-files /var/lib/layerfiles/repo-files >/var/log/fuse-filewatcher.log 2>&1& sleep 1'", vmI.GetHostIP()),
	)
	if err != nil {
		return errors.Wrap(err, "could not start filewatcher client")
	}
	return nil
}

func (server *FileShareServer) Start(vm *vm.QemuVM) error {
	if server.running {
		return nil
	}
	err := server.ensureFuseClientRunning(vm)
	if err != nil {
		return err
	}
	server.fuseServer = fusefilewatcherserver.NewFuseFilewatcherServer()
	server.fuseServer.RPCListenAddr = ":30811"
	server.fuseServer.MetaHost = "localhost:30812"
	server.fuseServer.OnRead = func(path string) {
		log.Println("File read: ", path)
	}
	go func() {
		log.Fatal(server.fuseServer.Run())
	}()
	server.running = true
	return nil
}

func (server *FileShareServer) Copy(ctx context.Context, source string) error {
	conn := server.fuseServer.WaitForConn()
	if conn == nil {
		return fmt.Errorf("connection never finished")
	}
	res, err := conn.Copy(ctx, &filewatcher_model.CopyReq{HostSource: source})
	if err != nil {
		return err
	}
	if res.Error != "" {
		return fmt.Errorf("%v", res.Error)
	}
	return nil
}