package instructions

import (
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"github.com/pkg/errors"
	"hash"
)

type User struct {
	Username string
}

func ParseUserInstruction(stream *tokenstream.TokenStream) (*User, error) {
	if !stream.HasToken() {
		return nil, errors.New("USER instruction was missing a username.")
	}

	tok := stream.Pop()
	if tok.GetTokenType() != lexer.LayerfileUSER_NAME {
		return nil, fmt.Errorf("invalid token: %v", tok.GetText())
	}

	return &User{Username: tok.GetText()}, nil
}

func (user *User) String() string {
	return fmt.Sprintf("USER %s", user.Username)
}

func (user *User) Hash(h hash.Hash, context *hashcontext.HashContext) {
	h.Write([]byte(user.String()))
	h.Write([]byte{0})
}
