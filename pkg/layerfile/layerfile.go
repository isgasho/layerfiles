package layerfile

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/instructions"
	"sort"
)

type Layerfile struct {
	//ImageId : e.g., vm/ubuntu:18.04, snapshot/10, ...
	ImageId      string
	Instructions []instructions.Instruction
	FilePath     string
}

func CalculateHashes(instrs []instructions.Instruction, context *hashcontext.HashContext) []string {
	hasher := sha1.New()
	hasher.Write([]byte(context.Image))
	hasher.Write([]byte{0})
	if context.NumCPUs != 0 {
		hasher.Write([]byte(fmt.Sprint(context.NumCPUs)))
		hasher.Write([]byte{0})
	}

	sort.Strings(context.SecretEnv)

	instrHashes := make([]string, len(instrs))
	for i, instr := range instrs {
		instr.Hash(hasher, context)
		instrHashes[i] = hex.EncodeToString(hasher.Sum(nil))
	}
	return instrHashes
}
