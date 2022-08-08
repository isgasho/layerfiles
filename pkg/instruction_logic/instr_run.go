package instruction_logic

import (
	"github.com/webappio/layerfiles/pkg/layerfile/instructions"
	"github.com/webappio/layerfiles/pkg/util"
	"strings"
)

func (runner *InstructionRunner) ProcessRunCommand(cmd *instructions.Run) error {
	commandHandler := runner.VM.GetCommandHandler()

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
	return commandHandler.RunCommandStreamOutput(rootCommand)
}