package instructions

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
	"strings"
)

type SetupFile struct {
	Files []string
}

func ParseSetupFileInstruction(stream *tokenstream.TokenStream) (*SetupFile, error) {
	files, err := parseFiles(stream)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse 'SETUP FILE' instruction")
	}
	if len(files) < 1 {
		return nil, fmt.Errorf("usage: SETUP FILE [file...]")
	}
	return &SetupFile{Files: files}, nil
}

func (setup *SetupFile) String() string {
	return fmt.Sprintf("SETUP FILE %s", strings.Join(setup.Files, " "))
}

func (setup *SetupFile) Hash(h hash.Hash, context *hashcontext.HashContext) {
	h.Write([]byte(setup.String()))
	h.Write([]byte{0})
}
