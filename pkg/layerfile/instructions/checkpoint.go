package instructions

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
)

type Checkpoint struct {
	Name string
}

func ParseCheckpointInstruction(stream *tokenstream.TokenStream) (*Checkpoint, error) {
	res := &Checkpoint{}

	for token := stream.Pop(); token != nil && token.GetTokenType() != lexer.LayerfileCHECKPOINT_EOL; token = stream.Pop() {
		if token.GetTokenType() == lexer.LayerfileCHECKPOINT_VALUE {
			if res.Name != "" {
				return nil, errors.New("checkpoint had two names specified")
			}
			res.Name = token.GetText()
		} else {
			return nil, errors.New("Unexpected token type for " + token.GetText())
		}
	}
	return res, nil
}

func (checkpoint *Checkpoint) String() string {
	if checkpoint.Name != "" {
		return fmt.Sprintf("CHECKPOINT %s", checkpoint.Name)
	} else {
		return "CHECKPOINT"
	}
}

func (checkpoint *Checkpoint) Hash(h hash.Hash, context *hashcontext.HashContext) {
	if checkpoint.Name == "disabled" {
		h.Write([]byte{0})
		h.Write([]byte("CHECKPOINT disabled"))
	}
	//checkpoints do not affect the hash
}
