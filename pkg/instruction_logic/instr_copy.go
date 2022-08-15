package instruction_logic

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/layerfile/instructions"
	"github.com/webappio/layerfiles/pkg/util"
	"log"
	"math/rand"
	"path/filepath"
	"strings"
)

func (runner *InstructionRunner) ProcessCopyCommand(cmd *instructions.Copy) error {
	err := runner.VMContactServer.Start(runner.VM)
	if err != nil {
		return err
	}

	if len(cmd.SourceFiles) != 1 {
		return fmt.Errorf("only one copy source is supported for now")
	}

	if strings.HasPrefix(cmd.TargetFile, "~/") {
		return fmt.Errorf("home directory COPY destinations are not supported yet") //TODO
	}

	if strings.HasPrefix(cmd.SourceFiles[0], "~/") {
		return fmt.Errorf("COPY source cannot start with ~/")
	}

	absSource := ""
	if filepath.IsAbs(cmd.SourceFiles[0]) {
		absSource = cmd.SourceFiles[0]
	} else {
		absSource, err = filepath.Abs(filepath.Join(filepath.Dir(runner.Layerfile.FilePath), cmd.SourceFiles[0]))
		if err != nil {
			return fmt.Errorf("could not get absolute path for copy: %v", cmd.SourceFiles[0])
		}
	}

	err = runner.VMContactServer.Copy(context.TODO(), absSource)
	if err != nil {
		out, _ := runner.VM.GetCommandHandler().RunCommand("cat /var/log/fuse-filewatcher.log")
		if out != "" {
			log.Println(out)
		}
		return err
	}

	//TODO use unique id instead of random
	workDir := fmt.Sprintf("/var/lib/layerfiles/cp-mount-%d-work", rand.Intn(1000000))
	lowerDir := fmt.Sprintf("/var/lib/layerfiles/repo-files/%v", filepath.Base(absSource))
	cleanTarget := cmd.TargetFile
	if cleanTarget == "." {
		cleanTarget = "/root"
	} else if strings.HasPrefix(cleanTarget, "./") {
		cleanTarget = "/root/" + cleanTarget[2:]
	}
	cleanTarget = util.SanitizeBashPath(cleanTarget)

	log.Println(workDir, ", ", cleanTarget, ", ", lowerDir)
	_, err = runner.VM.GetCommandHandler().RunCommand("" +
		"mkdir -p " + util.SanitizeBashPath(workDir) +
		" && mkdir -p " + cleanTarget +
		" && mount -t overlay overlay" +
		" -o lowerdir=" + util.SanitizeBashPath(lowerDir) +
		",upperdir=" + cleanTarget +
		",workdir=" + util.SanitizeBashPath(workDir) +
		" " + cleanTarget +
		" -vvvv")
	if err != nil {
		return errors.Wrapf(err, "error creating mount for COPY directive")
	}

	return nil
}