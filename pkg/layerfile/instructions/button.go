package instructions

import (
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
	"strings"
)

type Button struct {
	Message string
}

func ParseButtonInstruction(stream *tokenstream.TokenStream) (*Button, error) {
	token := stream.Pop()
	if token == nil {
		return nil, fmt.Errorf("BUTTON requires a message")
	}
	if token.GetTokenType() != lexer.LayerfileBUTTON_DATA {
		return nil, fmt.Errorf("unexpected token while reading 'BUTTON': %s", token.GetText())
	}
	message := strings.TrimSpace(token.GetText())
	return &Button{Message: message}, nil
}

func (button *Button) String() string {
	return "BUTTON " + button.Message
}

func (button *Button) Hash(hash.Hash, *hashcontext.HashContext) {
	//do nothing - buttons do not edit the VM
}
