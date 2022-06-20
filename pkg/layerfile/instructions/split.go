package instructions

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
	"strconv"
	"strings"
)

type Split struct {
	Count int
}

func ParseSplitInstruction(stream *tokenstream.TokenStream) (*Split, error) {
	token := stream.Pop()
	if token == nil {
		return nil, fmt.Errorf("expected argument after 'SPLIT', got end of file")
	}
	if token.GetTokenType() != lexer.LayerfileSPLIT_NUMBER {
		return nil, fmt.Errorf("unexpected token while reading 'SPLIT': %s", token.GetText())
	}
	target, err := strconv.ParseInt(strings.TrimSpace(token.GetText()), 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "usage is SPLIT (number)")
	}

	return &Split{Count: int(target)}, nil
}

func (split *Split) String() string {
	return fmt.Sprintf("SPLIT %d", split.Count)
}

func (split *Split) Hash(h hash.Hash, context *hashcontext.HashContext) {
	h.Write([]byte(split.String()))
	h.Write([]byte{0})
}
