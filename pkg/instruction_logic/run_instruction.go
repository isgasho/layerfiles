package instruction_logic

import (
	"fmt"
	file_share_server "github.com/webappio/layerfiles/pkg/file-share-server"
	"github.com/webappio/layerfiles/pkg/layerfile"
	"github.com/webappio/layerfiles/pkg/layerfile/instructions"
	"github.com/webappio/layerfiles/pkg/vm"
	"regexp"
)

var loginRegex = regexp.MustCompile("login:")
var passwordRegex = regexp.MustCompile("Password:")
var promptRegex = regexp.MustCompile("root@ubuntu2204-layerfile:.*")
var doneRegex = regexp.MustCompile("done.*?\n")

type InstructionRunner struct {
	VM *vm.QemuVM
	Layerfile *layerfile.Layerfile
	FileShareServer *file_share_server.FileShareServer
}

func (runner *InstructionRunner) RunInstructions() error {
	commandHandler := runner.VM.GetCommandHandler()
	commandHandler.WaitForRegex(loginRegex, nil)
	commandHandler.Stdin.Write([]byte("root\n"))
	commandHandler.WaitForRegex(passwordRegex, nil)
	commandHandler.Stdin.Write([]byte("password\n"))
	commandHandler.WaitForRegex(promptRegex, nil)
	commandHandler.Stdin.Write([]byte("export PROMPT_COMMAND='PS1=\"\"'; stty -echo; mkdir -p /var/lib/layerfiles; echo 'd'one\n"))
	commandHandler.WaitForRegex(doneRegex, nil)

	runner.FileShareServer = &file_share_server.FileShareServer{}

	for _, instr := range runner.Layerfile.Instructions {
		fmt.Println(instr)
		switch instr := instr.(type) {
		case *instructions.Run:
			err := runner.ProcessRunCommand(instr)
			if err != nil {
				return err
			}
		case *instructions.Copy:
			err := runner.ProcessCopyCommand(instr)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("cannot yet process instruction of type %T", instr)
		}
	}
	return nil
}