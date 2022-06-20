package instructions

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
	"strings"
)

type SecretEnv struct {
	Secrets []string
}

func ParseSecretEnvInstruction(stream *tokenstream.TokenStream) (*SecretEnv, error) {
	secrets := []string{}

	for token := stream.Pop(); token != nil && token.GetTokenType() != lexer.LayerfileSECRET_ENV_EOL; token = stream.Pop() {
		if token.GetTokenType() == lexer.LayerfileSECRET_ENV_VALUE {
			secrets = append(secrets, token.GetText())
		} else {
			return nil, errors.New("Unexpected token type for " + token.GetText())
		}
	}
	if len(secrets) == 0 {
		return nil, fmt.Errorf("SECRET ENV must be followed by at least one secret name")
	}
	return &SecretEnv{Secrets: secrets}, nil
}

func (secretEnv *SecretEnv) String() string {
	return fmt.Sprintf("SECRET ENV %s", strings.Join(secretEnv.Secrets, " "))
}

func (secretEnv *SecretEnv) Hash(h hash.Hash, context *hashcontext.HashContext) {
	if context.SecretEnv != nil {
		for _, secret := range secretEnv.Secrets { //SECRET ENV a b c
			for _, secretVal := range context.SecretEnv { //a=b c=d e=f
				if strings.HasPrefix(secretVal, secret+"=") {
					h.Write([]byte(secretVal))
				}
			}
		}
	}
	h.Write([]byte(secretEnv.String()))
	h.Write([]byte{0})
}
