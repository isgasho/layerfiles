package main

import (
	"flag"
	"fmt"
	"github.com/webappio/layerfiles/pkg/vm_protocol/client"
	"k8s.io/klog"
	"os"
)

func main() {
	logFlagSet := flag.NewFlagSet("Logging flags", flag.ExitOnError)
	klog.InitFlags(logFlagSet)
	_ = logFlagSet.Parse(os.Args[1:])

	if logFlagSet.NArg() != 4 {
		fmt.Printf("Usage: %v [meta host] [rpc host] [src] [dest]\n", os.Args[0])
		fmt.Println("E.g., localhost:3214 localhost:3215 /src /dest")
		return
	}

	klog.Info("Starting vm_protocol!")
	(&client.FuseFilewatcherClient{
		MetaListenAddr: logFlagSet.Arg(0),
		RPCHost:        logFlagSet.Arg(1),
		Source:         logFlagSet.Arg(2),
		Dest:           logFlagSet.Arg(3),
	}).Run()
}
