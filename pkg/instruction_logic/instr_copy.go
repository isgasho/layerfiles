package instruction_logic

import (
	"context"
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/instructions"
	"log"
	"path/filepath"
)

func (runner *InstructionRunner) ProcessCopyCommand(cmd *instructions.Copy) error {
	err := runner.FileShareServer.Start(runner.VM)
	if err != nil {
		return err
	}

	if len(cmd.SourceFiles) != 1 {
		return fmt.Errorf("only one copy source is supported for now")
	}

	absPath := ""
	if filepath.IsAbs(cmd.SourceFiles[0]) {
		absPath = cmd.SourceFiles[0]
	} else {
		absPath, err = filepath.Abs(filepath.Join(filepath.Dir(runner.Layerfile.FilePath), cmd.SourceFiles[0]))
		if err != nil {
			return fmt.Errorf("could not get absolute path for copy: %v", cmd.SourceFiles[0])
		}
	}

	err = runner.FileShareServer.Copy(context.TODO(), absPath)
	if err != nil {
		out, _ := runner.VM.GetCommandHandler().RunCommand("cat /var/log/fuse-filewatcher.log")
		if out != "" {
			log.Println(out)
		}
		return err
	}
	return nil
}