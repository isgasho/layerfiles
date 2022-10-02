package vm_contact

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/vm"
)

type VMContactServer struct {
	running bool
}

func (server *VMContactServer) ensureFuseClientRunning(vmI *vm.QemuVM) error {
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

func (server *VMContactServer) Start(vm *vm.QemuVM) error {
	if server.running {
		return nil
	}
	err := server.ensureFuseClientRunning(vm)
	if err != nil {
		return err
	}
	//server.srv = vm_protocol_server.NewFuseFilewatcherServer()
	//server.srv.RPCListenAddr = ":30811"
	//server.srv.MetaHost = "localhost:30812"
	//server.srv.OnRead = func(path string) {
	//	log.Println("File read: ", path)
	//}
	//go func() {
	//	log.Fatal(server.srv.Run())
	//}()
	//server.running = true
	return nil
}

func (server *VMContactServer) Copy(ctx context.Context, source string) error {
	//conn := server.srv.WaitForConn()
	//if conn == nil {
	//	return fmt.Errorf("connection never finished")
	//}
	//res, err := conn.Copy(ctx, &vm_protocol_model.CopyReq{HostSource: source})
	//if err != nil {
	//	return err
	//}
	//if res.Error != "" {
	//	return fmt.Errorf("%v", res.Error)
	//}
	//
	//_, err = conn.AllowReads(ctx, &vm_protocol_model.AllowReadsReq{})
	//if err != nil {
	//	return err
	//}
	return nil
}

func (server *VMContactServer) ExposeWebsite(scheme string, dom string, port uint32, path string, rewritePath string) (domain string, err error) {
	//res, err := server.srv.WaitForConn().ExposeWebsite(context.Background(), &vm_protocol_model.ExposeWebsiteReq{
	//	IsHttps:     scheme == "https",
	//	Domain:      dom,
	//	Port:        port,
	//	Path:        path,
	//	RewritePath: rewritePath,
	//})
	//if err != nil {
	//	return "", err
	//}
	//return res.Host, nil
	return "", nil
}
