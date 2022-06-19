package instructions

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
	"strings"
)

type RunType byte

const RunTypeStandard = RunType(0)
const RunTypeBackground = RunType(1)
const RunTypeRepeatable = RunType(2)

type Run struct {
	Command string
	Type    RunType
}

func ParseRunInstruction(stream *tokenstream.TokenStream, runToken antlr.Token) (*Run, error) {
	token := stream.Pop()
	if token == nil {
		return nil, fmt.Errorf("expected argument after 'RUN', got end of file")
	}
	if token.GetTokenType() != lexer.LayerfileRUN_DATA {
		return nil, fmt.Errorf("unexpected token while reading 'RUN': %s", token.GetText())
	}
	command := strings.TrimSpace(token.GetText())

	runCmd := &Run{Command: command}
	switch runToken.GetTokenType() {
	case lexer.LayerfileRUN_BACKGROUND:
		runCmd.Type = RunTypeBackground
	case lexer.LayerfileRUN_REPEATABLE:
		runCmd.Type = RunTypeRepeatable
	case lexer.LayerfileRUN:
		runCmd.Type = RunTypeStandard
	default:
		return nil, fmt.Errorf("unknown RUN command: '%v'", runToken.GetText())
	}
	return runCmd, nil
}

func (run *Run) String() string {
	var op string
	switch run.Type {
	case RunTypeStandard:
		op = "RUN"
	case RunTypeBackground:
		op = "RUN BACKGROUND"
	case RunTypeRepeatable:
		op = "RUN REPEATABLE"
	}
	return fmt.Sprintf("%s %s", op, run.Command)
}

func (run *Run) Hash(h hash.Hash, context *hashcontext.HashContext) {
	h.Write([]byte{byte(run.Type)})
	h.Write([]byte(run.String()))
	h.Write([]byte{0})
}
