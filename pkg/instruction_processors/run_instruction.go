package instruction_processors

import (
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile"
	"github.com/webappio/layerfiles/pkg/layerfile/instructions"
	"github.com/webappio/layerfiles/pkg/vm"
	"regexp"
)

var loginRegex = regexp.MustCompile("login:")
var passwordRegex = regexp.MustCompile("Password:")
var promptRegex = regexp.MustCompile("root@ubuntu2204-layerfile:.*")
func RunInstructions(vm *vm.QemuVM, lf *layerfile.Layerfile) error {
	commandHandler := vm.GetCommandHandler()
	commandHandler.WaitForRegex(loginRegex, true)
	commandHandler.Stdin.Write([]byte("root\n"))
	commandHandler.WaitForRegex(passwordRegex, true)
	commandHandler.Stdin.Write([]byte("password\n"))
	commandHandler.WaitForRegex(promptRegex, true)
	for _, instr := range lf.Instructions {
		switch instr := instr.(type) {
		case *instructions.Run:
			commandHandler.Stdin.Write([]byte(instr.Command + "\n"))
			commandHandler.WaitForRegex(promptRegex, true)
		default:
			return fmt.Errorf("cannot yet process instruction of type %T", instr)
		}
	}
	return nil
}