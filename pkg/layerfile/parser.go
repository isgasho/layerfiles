package layerfile

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/layerfile/continuationreader"
	"github.com/webappio/layerfiles/pkg/layerfile/instructions"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"io"
	"strings"
)

type lexerErrorListener struct {
	*antlr.DefaultErrorListener

	lineErrors map[int]error
}

type ParseError struct {
	Line    int
	Column  int
	Message string
}

func (err *ParseError) Error() string {
	return fmt.Sprintf("syntax error at line %d: %s", err.Line+1, err.Message)
}

func (listener *lexerErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	if _, ok := listener.lineErrors[line]; !ok {
		listener.lineErrors[line] = &ParseError{Line: line, Column: column, Message: msg}
	}
}

func tokenizeLayerfile(reader io.Reader) (*tokenstream.TokenStream, error) {
	input, err := continuationreader.New(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading Layerfile: %s", err.Error())
	}
	lex := lexer.NewLayerfile(input)

	lex.RemoveErrorListeners()
	errListener := &lexerErrorListener{
		lineErrors: make(map[int]error),
	}
	lex.AddErrorListener(errListener)

	tokenStream := antlr.NewCommonTokenStream(lex, 0)
	tokenStream.Fill()

	ret, err := tokenstream.Read(tokenStream)
	if err != nil {
		return nil, errors.Wrap(err, "could not read tokens in layerfile")
	}

	for _, err := range errListener.lineErrors {
		return nil, err
	}
	return ret, nil
}

func parseInstruction(token antlr.Token, stream *tokenstream.TokenStream) (instructions.Instruction, error) {
	switch token.GetTokenType() {
	case lexer.LayerfileBUTTON:
		return instructions.ParseButtonInstruction(stream)
	case lexer.LayerfileBUILD_ENV:
		return instructions.ParseBuildEnvInstruction(stream)
	case lexer.LayerfileCHECKPOINT:
		return instructions.ParseCheckpointInstruction(stream)
	case lexer.LayerfileCLONE:
		return instructions.ParseCloneInstruction(stream)
	case lexer.LayerfileCOPY:
		return instructions.ParseCopyInstruction(stream)
	case lexer.LayerfileEXPOSE_TCP:
		return instructions.ParseExposeTcpInstruction(stream)
	case lexer.LayerfileEXPOSE_WEBSITE:
		return instructions.ParseExposeWebsiteInstruction(stream)
	case lexer.LayerfileENV:
		return instructions.ParseEnvInstruction(stream)
	case lexer.LayerfileMEMORY:
		return instructions.ParseMemoryInstruction(stream)
	case lexer.LayerfileRUN, lexer.LayerfileRUN_BACKGROUND, lexer.LayerfileRUN_REPEATABLE:
		return instructions.ParseRunInstruction(stream, token)
	case lexer.LayerfileSECRET_ENV:
		return instructions.ParseSecretEnvInstruction(stream)
	case lexer.LayerfileSETUP_FILE:
		return instructions.ParseSetupFileInstruction(stream)
	case lexer.LayerfileSKIP_REMAINING_IF:
		return instructions.ParseSkipRemainingIfInstruction(stream)
	case lexer.LayerfileSPLIT:
		return instructions.ParseSplitInstruction(stream)
	case lexer.LayerfileFROM:
		return instructions.ParseFromInstruction(stream)
	case lexer.LayerfileCACHE:
		return instructions.ParseCacheInstruction(stream)
	case lexer.LayerfileUSER:
		return instructions.ParseUserInstruction(stream)
	case lexer.LayerfileWAIT:
		return instructions.ParseWaitInstruction(stream)
	case lexer.LayerfileWORKDIR:
		return instructions.ParseWorkdirInstruction(stream)
	case lexer.LayerfileLABEL:
		return instructions.ParseLabelInstruction(stream)
	case lexer.LayerfileAWS:
		return instructions.ParseAWSInstruction(stream)
	default:
		return nil, fmt.Errorf("unknown instruction %s at %d:%d", token.GetText(), token.GetLine(), token.GetColumn())
	}
}

func parseInstructions(stream *tokenstream.TokenStream) ([]instructions.Instruction, error) {
	instrs := []instructions.Instruction{}

	for stream.HasToken() {
		token := stream.Pop()

		instr, err := parseInstruction(token, stream)
		if err != nil {
			return nil, err
		}

		instrs = append(instrs, instr)
	}

	return instrs, nil
}

func parseFrom(instrs []instructions.Instruction) (string, []instructions.Instruction, error) {
	var fromImage string
	if fromInstr, ok := instrs[0].(*instructions.From); ok {
		fromImage = fromInstr.ImageId
	} else {
		return "", nil, errors.New("first instruction in Layerfile must be 'FROM'")
	}
	return fromImage, instrs[1:], nil
}

func ParseInstruction(text string) (instructions.Instruction, error) {
	tokenStream, err := tokenizeLayerfile(strings.NewReader(text))
	if err != nil {
		return nil, errors.Wrap(err, "error while tokenizing instruction")
	}
	instrs, err := parseInstructions(tokenStream)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse that instruction")
	}
	if len(instrs) != 1 {
		return nil, errors.Wrap(err, "there were multiple instructions on this line")
	}
	return instrs[0], nil
}

func ReadLayerfile(contents string) (*Layerfile, error) {
	tokenStream, err := tokenizeLayerfile(strings.NewReader(contents))
	if err != nil {
		return nil, errors.Wrap(err, "error while tokenizing Layerfile")
	}

	instrs, err := parseInstructions(tokenStream)
	if err != nil {
		return nil, errors.Wrap(err, "error while parsing Layerfile groups")
	}

	if len(instrs) == 0 {
		return nil, errors.New("empty layerfile")
	}

	image, instrs, err := parseFrom(instrs)
	if err != nil {
		return nil, err
	}

	return &Layerfile{
		ImageId:      image,
		Instructions: instrs,
	}, nil
}
