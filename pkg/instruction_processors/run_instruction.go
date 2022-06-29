package instruction_processors

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/layerfile"
	"github.com/webappio/layerfiles/pkg/layerfile/instructions"
	"github.com/webappio/layerfiles/pkg/util"
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

	userCommand := `touch /var/lib/layerfiles/cwd; cd "$(cat /var/lib/layerfiles/cwd)" ;`
	userCommand += "touch ~/.profile; touch ~/.bash_profile ; source /etc/profile; source ~/.profile; source ~/.bash_profile;"
	screenName := "instruction_1" //TODO _1 hardcoded
	screenLog := "/var/log/" + screenName + ".log"
	if cmd.Type == instructions.RunTypeBackground {
		userCommand += strings.TrimSpace(cmd.Command)
		userCommand += `; echo -e "RGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu $? nRGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu" >` + screenLog
	} else {
		userCommand += cmd.Command
	}

	//this chunk runs as root
	rootCommand := ""
	rootCommand += `mkdir -p "$(cat /var/lib/layerfiles/cwd 2>/dev/null || echo /root)" ;`
	rootCommand += "touch /etc/profile ; chmod 644 /etc/profile ;"
	rootCommand += "if [ ! -f /var/lib/layerfiles/curr-user ]; then echo root > /var/lib/layerfiles/curr-user; fi ;"
	rootCommand += "if [ ! -f /var/lib/layerfiles/cwd ]; then echo /root > /var/lib/layerfiles/cwd; fi ;"

	shell := "/bin/bash"
	if cmd.Type == instructions.RunTypeBackground {
		shell = "/usr/bin/env screen -S " + screenName + ` -Logfile "` + screenLog + `.tmp" -mdL bash`
	}

	rootCommand += "sudo -u $(cat /var/lib/layerfiles/curr-user) -H -- " + shell + " -c " + util.SanitizeBashPath(userCommand) + " ;"
	if cmd.Type == instructions.RunTypeBackground {
		rootCommand += "sleep 2; "
		rootCommand += `[ $(grep --no-messages -oP '(?<=RGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu\s)(\d*)(?=\snRGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu)' `
		rootCommand += screenLog + " || echo 0) == 0 ]; "
	}
	rootCommand += "echo -e '\\n\\nRGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu' $? 'nRGlzdHJpYnV0ZWQgQ29udGFpbmVycyBJbmMu' "

	//if len(userCommand) > 64 {
	//	fmt.Println(userCommand[:64])
	//} else {
	//	fmt.Println(userCommand)
	//}

	commandHandler.Stdin.Write([]byte(rootCommand + "\n"))
	match, err := commandHandler.WaitForRegex(commandDoneStatusRegex, true)
	if err != nil {
		return err
	}

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
	commandHandler.Stdin.Write([]byte("export PROMPT_COMMAND='PS1=\"\"'; stty -echo; mkdir -p /var/lib/layerfiles; echo 'd'one\n"))
	commandHandler.WaitForRegex(doneRegex, false)

	for _, instr := range lf.Instructions {
		fmt.Println("Running", instr)
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