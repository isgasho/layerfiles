package instructions

import (
	"bytes"
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
	"strings"
)

type Env struct {
	Env []string
}

func ParseEnvInstruction(stream *tokenstream.TokenStream) (*Env, error) {
	env := []string{}

	prevToken := ""
	for token := stream.Pop(); token != nil && token.GetTokenType() != lexer.LayerfileENV_EOL; token = stream.Pop() {
		tokenValue := token.GetText()
		if strings.Contains(tokenValue, "=") {
			env = append(env, tokenValue)
		} else if prevToken != "" {
			env = append(env, prevToken+"="+tokenValue)
			prevToken = ""
		} else {
			prevToken = tokenValue
		}
	}
	if len(env) == 0 {
		return nil, fmt.Errorf("ENV must be followed by at least one var=value pair")
	}
	return &Env{Env: env}, nil
}

func (env *Env) String() string {
	var buf bytes.Buffer
	buf.WriteString("ENV")
	for _, val := range env.Env {
		buf.WriteRune(' ')
		buf.WriteString(val)
	}
	return buf.String()
}

func (env *Env) Hash(h hash.Hash, context *hashcontext.HashContext) {
	h.Write([]byte(env.String()))
	h.Write([]byte{0})
}
