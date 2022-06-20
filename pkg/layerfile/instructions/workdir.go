package instructions

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
)

type Workdir struct {
	Dir string
}

func ParseWorkdirInstruction(stream *tokenstream.TokenStream) (*Workdir, error) {
	dirs, err := parseFiles(stream)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse 'WORKDIR' instruction")
	}
	if len(dirs) != 1 {
		return nil, fmt.Errorf("usage: WORKDIR [directory]")
	}
	return &Workdir{Dir: dirs[0]}, nil
}

func (wd *Workdir) String() string {
	return fmt.Sprintf("WORKDIR %s", wd.Dir)
}

func (wd *Workdir) Hash(h hash.Hash, context *hashcontext.HashContext) {
	h.Write([]byte(wd.String()))
	h.Write([]byte{0})
}
