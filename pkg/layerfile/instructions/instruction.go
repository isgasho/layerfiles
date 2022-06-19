package instructions

import (
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
)

type Instruction interface {
	//Hash writes this instruction's hash to the given hasher
	Hash(dest hash.Hash, context *hashcontext.HashContext)
}

func parseFiles(stream *tokenstream.TokenStream) ([]string, error) {
	files := []string{}
	for token := stream.Pop(); token != nil && token.GetTokenType() != lexer.LayerfileEND_OF_FILES; token = stream.Pop() {
		if token.GetTokenType() == lexer.LayerfileFILE {
			files = append(files, token.GetText())
		} else {
			return nil, fmt.Errorf("unexpected token type while reading file list: %s", token.GetText())
		}
	}

	return files, nil
}
