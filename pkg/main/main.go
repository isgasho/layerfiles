package main

import (
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile_graph"
	"github.com/webappio/layerfiles/pkg/vm"
	"log"
	"os"
	"os/signal"
)

func main() {
	layerfiles, err := layerfile_graph.FindLayerfiles(".")
	if err != nil {
		log.Fatal(err)
	}
	if len(layerfiles) != 1 {
		fmt.Println("You must have exactly one Layerfile in your repository.")
		os.Exit(1)
	}

	qemuVM := vm.QemuVM{}
	qemuVM.Start()

	sigHandler := make(chan os.Signal, 2)
	signal.Notify(sigHandler, os.Interrupt)
	<-sigHandler
	qemuVM.Stop()
}
