package instruction_logic

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/layerfile/instructions"
)

func (runner *InstructionRunner) ProcessExposeWebsiteCommand(cmd *instructions.ExposeWebsite) error {
	dom, err := runner.VMContactServer.ExposeWebsite(cmd.Scheme, cmd.Domain, uint32(cmd.Port), cmd.Path, cmd.RewritePath)
	if err != nil {
		return errors.Wrap(err, "could not expose the website - is your internet working?")
	}
	fmt.Println("EXPOSE WEBSITE is available at https://" + dom)
	return nil
}
