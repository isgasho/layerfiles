package instructions

import (
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
	"strings"
)

type Wait struct {
	Targets []string
}

func ParseWaitInstruction(stream *tokenstream.TokenStream) (*Wait, error) {
	files, err := parseFiles(stream)
	if err != nil {
		return nil, err
	}

	return &Wait{Targets: files}, nil
}

func (wait *Wait) String() string {
	return fmt.Sprintf("WAIT %s", strings.Join(wait.Targets, " "))
}

func (wait *Wait) Hash(h hash.Hash, context *hashcontext.HashContext) {
	h.Write([]byte(wait.String()))
	h.Write([]byte{0})
}
