package commands

import (
	"flag"
	"fmt"
	"github.com/webappio/layerfiles/pkg/instruction_logic"
	"github.com/webappio/layerfiles/pkg/layerfile_graph"
	"github.com/webappio/layerfiles/pkg/vm"
	"log"
	"os"
	"os/signal"
)

var buildFlags = flag.NewFlagSet("build", flag.ExitOnError)

func Build(args []string) {
	buildFlags.ErrorHandling()
	buildFlags.Parse(args)

	layerfiles, err := layerfile_graph.FindLayerfiles(".")
	if err != nil {
		log.Fatal(err)
	}
	if len(layerfiles) != 1 {
		fmt.Println("You must have exactly one Layerfile in your repository.")
		os.Exit(1)
	}

	qemuVM := &vm.QemuVM{}
	err = qemuVM.Start()
	if err != nil {
		log.Fatal(err)
	}

	instructionsDone := make(chan interface{}, 1)
	go func() {
		err := (&instruction_logic.InstructionRunner{VM: qemuVM, Layerfile: layerfiles[0]}).RunInstructions()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		close(instructionsDone)
	}()

	sigHandler := make(chan os.Signal, 2)
	signal.Notify(sigHandler, os.Interrupt)
	select {
	case <-sigHandler:
		fmt.Println("Exiting.")
	case <-instructionsDone:
	}

	qemuVM.Stop()
}
