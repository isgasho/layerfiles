package instructions

import (
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"github.com/pkg/errors"
	"hash"
	"strings"
)

type Cache struct {
	Dirs []string
}

func ParseCacheInstruction(stream *tokenstream.TokenStream) (*Cache, error) {
	dirs, err := parseFiles(stream)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse 'CACHE' instruction")
	}
	if len(dirs) == 0 {
		return nil, fmt.Errorf("usage: CACHE [cache directories...]")
	}
	return &Cache{Dirs: dirs}, nil
}

func (cache *Cache) String() string {
	return fmt.Sprintf("CACHE %s", strings.Join(cache.Dirs, " "))
}

func (cache *Cache) Hash(h hash.Hash, context *hashcontext.HashContext) {
	h.Write([]byte(cache.String()))
	h.Write([]byte{0})
}
