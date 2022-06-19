package instructions

import (
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
	"strings"
)

type Clone struct {
	CloneURL      string
	DefaultBranch string
	Sources       []string
	Dest          string
}

func ParseCloneInstruction(stream *tokenstream.TokenStream) (*Clone, error) {
	res := &Clone{}

	tokens := []string{}
	for token := stream.Pop(); token != nil && token.GetTokenType() != lexer.LayerfileCLONE_EOL; token = stream.Pop() {
		if strings.HasPrefix(token.GetText(), "DEFAULT=") {
			res.DefaultBranch = token.GetText()[len("DEFAULT="):]
		} else {
			tokens = append(tokens, token.GetText())
		}
	}

	if len(tokens) < 2 {
		return nil, fmt.Errorf("CLONE invalid, format is CLONE (DEFAULT=master) [url] (source...) [destination]")
	}

	res.CloneURL = tokens[0]
	res.Dest = tokens[len(tokens)-1]
	res.Sources = tokens[1 : len(tokens)-1]

	return res, nil
}

func (clone *Clone) String() string {
	tokens := []string{}
	if clone.DefaultBranch != "" {
		tokens = append(tokens, "DEFAULT="+clone.DefaultBranch)
	}
	tokens = append(tokens, clone.CloneURL)
	tokens = append(tokens, clone.Sources...)
	tokens = append(tokens, clone.Dest)
	return fmt.Sprintf("CLONE %v", strings.Join(tokens, " "))
}

func (clone *Clone) Hash(h hash.Hash, context *hashcontext.HashContext) {
	h.Write([]byte(clone.String()))
	h.Write([]byte{0})
}
