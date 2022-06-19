package instructions

import (
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
	"strings"
)

type From struct {
	ImageId string
}

func ParseFromInstruction(stream *tokenstream.TokenStream) (*From, error) {
	token := stream.Pop()
	if token == nil {
		return nil, fmt.Errorf("FROM requires an image name after it")
	}
	if token.GetTokenType() != lexer.LayerfileFROM_DATA {
		return nil, fmt.Errorf("unexpected token while reading 'FROM': %s", token.GetText())
	}
	imageId := strings.TrimSpace(token.GetText())
	return &From{ImageId: imageId}, nil
}

func (from *From) String() string {
	return fmt.Sprintf("FROM %s", from.ImageId)
}

func (from *From) Hash(h hash.Hash, context *hashcontext.HashContext) {
	h.Write([]byte(from.String()))
	h.Write([]byte{0})
}
