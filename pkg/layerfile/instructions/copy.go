package instructions

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
	"strings"
)

type Copy struct {
	SourceFiles []string
	TargetFile  string
}

func ParseCopyInstruction(stream *tokenstream.TokenStream) (*Copy, error) {
	dirs, err := parseFiles(stream)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse 'COPY' instruction")
	}
	if len(dirs) < 2 {
		return nil, fmt.Errorf("usage: COPY [src...] [target]")
	}
	if len(dirs) >= 3 && !strings.HasSuffix(dirs[len(dirs)-1], "/") {
		return nil, fmt.Errorf("if copying multiple files, the destination must end with '/'")
	}
	return &Copy{SourceFiles: dirs[:len(dirs)-1], TargetFile: dirs[len(dirs)-1]}, nil
}

func (copy *Copy) String() string {
	return fmt.Sprintf("COPY %s %s", strings.Join(copy.SourceFiles, " "), copy.TargetFile)
}

func (copy *Copy) Hash(h hash.Hash, context *hashcontext.HashContext) {
	h.Write([]byte(copy.String()))
	h.Write([]byte{0})
}
