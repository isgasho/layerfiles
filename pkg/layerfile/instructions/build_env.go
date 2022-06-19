package instructions

import (
	"bytes"
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
)

type BuildEnv struct {
	BuildEnv []string
}

func ParseBuildEnvInstruction(stream *tokenstream.TokenStream) (*BuildEnv, error) {
	env := []string{}

	for token := stream.Pop(); token != nil && token.GetTokenType() != lexer.LayerfileBUILD_ENV_EOL; token = stream.Pop() {
		tokenValue := token.GetText()
		env = append(env, tokenValue)
	}
	if len(env) == 0 {
		return nil, fmt.Errorf("BUILD ENV must be followed by at least one value")
	}
	return &BuildEnv{BuildEnv: env}, nil
}

func (env *BuildEnv) String() string {
	var buf bytes.Buffer
	buf.WriteString("BUILD ENV")
	for _, val := range env.BuildEnv {
		buf.WriteRune(' ')
		buf.WriteString(val)
	}
	return buf.String()
}

func (env *BuildEnv) Hash(h hash.Hash, context *hashcontext.HashContext) {
	h.Write([]byte(env.String()))
	h.Write([]byte{0})
}
