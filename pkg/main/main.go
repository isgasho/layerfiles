package main

import (
	"flag"
	"fmt"
	"github.com/webappio/layerfiles/pkg/commands"
	"os"
	"strings"
)

var generalFlags = flag.NewFlagSet("", flag.ExitOnError)

var rootCommands = map[string]func(args []string){
	"build": commands.Build,
}

func printUsage() {
	fmt.Println("Usage:")
	allCommands := make([]string, 0, len(rootCommands))
	for cmd := range rootCommands {
		allCommands = append(allCommands, cmd)
	}
	fmt.Println(os.Args[0], "["+strings.Join(allCommands, "|")+"]")
}

func main() {
	generalFlags.Parse(os.Args[1:])

	if generalFlags.NArg() == 0 {
		printUsage()
		os.Exit(1)
	}

	generalFlags.Usage = printUsage

	cmd, ok := rootCommands[generalFlags.Arg(0)]
	if !ok {
		fmt.Println("Unknown command: '" + generalFlags.Arg(0) + "'")
		fmt.Println()
		printUsage()
		os.Exit(1)
	}

	cmd(generalFlags.Args()[1:])
}
