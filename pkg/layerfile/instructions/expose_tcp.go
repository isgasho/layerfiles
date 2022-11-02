package instructions

import (
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
	"regexp"
	"strconv"
)

type ExposeTcp struct {
	SourcePort   uint16
	DestPort uint16
}

var portPattern = regexp.MustCompile(`^:?(\d+)$`) // :8080 or 8080

func ParseExposeTcpInstruction(stream *tokenstream.TokenStream) (*ExposeTcp, error) {
	token := stream.Pop()
	if token == nil {
		return nil, fmt.Errorf("EXPOSE TCP requires a port immediately after it")
	}

	if token.GetTokenType() != lexer.LayerfileEXPOSE_TCP_ITEM {
		return nil, fmt.Errorf("unexpected token while reading 'EXPOSE TCP': %s", token.GetText())
	}

	match := portPattern.FindStringSubmatch(token.GetText())
	if match == nil {
		return nil, fmt.Errorf("EXPOSE TCP port was of incorrect form: %v", token.GetText())
	}

	sourcePort, err := strconv.ParseUint(match[1], 10, 16)
	if err != nil {
		return nil, err
	}

	if sourcePort == 0 {
		return nil, fmt.Errorf("Invalid EXPOSE TCP port: %v", sourcePort)
	}

	token = stream.Pop()
	if token == nil || token.GetTokenType() == lexer.LayerfileEXPOSE_TCP_EOL {
		return &ExposeTcp{SourcePort: uint16(sourcePort), DestPort: uint16(sourcePort)}, nil
	} else if token.GetTokenType() != lexer.LayerfileEXPOSE_TCP_ITEM {
		return nil, fmt.Errorf("invalid data in EXPOSE TCP instruction: %v", token.GetText())
	}

	match = portPattern.FindStringSubmatch(token.GetText())
	if match == nil {
		return nil, fmt.Errorf("EXPOSE TCP port was of incorrect form: %v", token.GetText())
	}

	destPort, err := strconv.ParseUint(match[1], 10, 16)
	if err != nil {
		return nil, err
	}

	if destPort == 0 {
		return nil, fmt.Errorf("Invalid EXPOSE TCP port: %v", destPort)
	}

	token = stream.Pop()
	if token == nil || token.GetTokenType() == lexer.LayerfileEXPOSE_TCP_EOL {
		return &ExposeTcp{SourcePort: uint16(sourcePort), DestPort: uint16(destPort)}, nil
	}
	return nil, fmt.Errorf("invalid data in EXPOSE TCP instruction: %v", token.GetText())
}

func (expose *ExposeTcp) String() string {
	return fmt.Sprintf("EXPOSE TCP :%v :%v", expose.SourcePort, expose.DestPort)
}

func (expose *ExposeTcp) Hash(h hash.Hash, context *hashcontext.HashContext) {
	h.Write([]byte{0})
	h.Write([]byte(expose.String()))
}
