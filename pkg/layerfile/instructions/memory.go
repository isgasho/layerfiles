package instructions

import (
	"encoding/binary"
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
	"regexp"
	"strconv"
	"strings"
)

type Memory struct {
	Amount int64
	Unit   string
}

func ParseMemoryInstruction(stream *tokenstream.TokenStream) (*Memory, error) {
	token := stream.Pop()
	if token == nil {
		return nil, fmt.Errorf("MEMORY requires an amount")
	}
	if token.GetTokenType() != lexer.LayerfileMEMORY_AMOUNT {
		return nil, fmt.Errorf("unexpected token while reading 'MEMORY': %s", token.GetText())
	}

	match := regexp.MustCompile("(\\d+)([gGmMkK]?)").FindStringSubmatch(token.GetText())
	if match == nil {
		return nil, fmt.Errorf("unexpected token while reading 'MEMORY': %s", token.GetText())
	}

	amount, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return nil, err
	}

	return &Memory{Amount: amount, Unit: strings.ToUpper(match[2])}, nil
}

func (memory *Memory) AmountBytes() int64 {
	bytes := memory.Amount
	switch strings.ToUpper(memory.Unit) {
	case "G":
		bytes *= 1024 * 1024 * 1024
	case "M":
		bytes *= 1024 * 1024
	case "K":
		bytes *= 1024
	case "":
	default:
		panic("invalid unit " + memory.Unit + " - this shouldn't happen!")
	}
	return bytes
}

func (memory *Memory) String() string {
	return fmt.Sprintf("MEMORY %d%s", memory.Amount, memory.Unit)
}

func (memory *Memory) Hash(hasher hash.Hash, context *hashcontext.HashContext) {
	hasher.Write([]byte("MEMORY"))
	var dest [8]byte
	binary.LittleEndian.PutUint64(dest[:], uint64(memory.AmountBytes()))
	hasher.Write(dest[:])
}
