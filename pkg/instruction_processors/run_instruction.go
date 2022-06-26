package instruction_processors

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/layerfile"
	"github.com/webappio/layerfiles/pkg/layerfile/instructions"
	"github.com/webappio/layerfiles/pkg/vm"
	"regexp"
	"strconv"
	"strings"
)

var loginRegex = regexp.MustCompile("login:")
var passwordRegex = regexp.MustCompile("Password:")
var promptRegex = regexp.MustCompile("root@ubuntu2204-layerfile:.*")
var doneRegex = regexp.MustCompile("done.*?\n")
var commandDoneStatusRegex = regexp.MustCompile("RGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu\\s+(\\d+)")

func ProcessRunCommand(vm *vm.QemuVM, cmd *instructions.Run) error {
	commandHandler := vm.GetCommandHandler()

	commandHandler.Stdin.Write([]byte(cmd.Command + "\n echo -e '\\n\\nRGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu' $? 'nRGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu'\n"))
	match, err := commandHandler.WaitForRegex(commandDoneStatusRegex, true)
	if err != nil {
		return err
	}

	//log.Printf("Command finished with status code %s", match[1])

	statusCode, err := strconv.ParseInt(strings.TrimSpace(string(match[1])), 10, 32)
	if err != nil {
		return errors.Wrap(err, "non-numeric exit status")
	}

	if statusCode != 0 {
		return fmt.Errorf("command failed with status code %d", statusCode)
	}
	return nil
}

func RunInstructions(vm *vm.QemuVM, lf *layerfile.Layerfile) error {
	commandHandler := vm.GetCommandHandler()
	commandHandler.WaitForRegex(loginRegex, false)
	commandHandler.Stdin.Write([]byte("root\n"))
	commandHandler.WaitForRegex(passwordRegex, false)
	commandHandler.Stdin.Write([]byte("password\n"))
	commandHandler.WaitForRegex(promptRegex, false)
	commandHandler.Stdin.Write([]byte("export PROMPT_COMMAND='PS1=\"\"'; stty -echo; echo done\n"))
	commandHandler.WaitForRegex(doneRegex, false)
	for _, instr := range lf.Instructions {
		switch instr := instr.(type) {
		case *instructions.Run:
			err := ProcessRunCommand(vm, instr)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("cannot yet process instruction of type %T", instr)
		}
	}
	return nil
}