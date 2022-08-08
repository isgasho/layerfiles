package main

import (
	"fmt"
	"github.com/webappio/layerfiles/pkg/fuse-filewatcher/server"
	"k8s.io/klog"
	"os"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Usage: %v [meta host] [rpc listen addr]\n", os.Args[0])
		fmt.Println("E.g., localhost:3214 localhost:3215 /repo-files")
		return
	}

	klog.Info("Starting fuse filewatcher server!")
	srv := server.NewFuseFilewatcherServer()
	srv.MetaHost = os.Args[1]
	srv.RPCListenAddr = os.Args[2]
	klog.Fatal(srv.Run())
}
