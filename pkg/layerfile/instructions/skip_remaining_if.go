package instructions

import (
	"bytes"
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
	"regexp"
	"strings"
)

type SkipRemainingIf struct {
	SkipRemainingIf []string
}

func ParseSkipRemainingIfInstruction(stream *tokenstream.TokenStream) (*SkipRemainingIf, error) {
	skipRemainingIf := []string{}

	for token := stream.Pop(); token != nil && token.GetTokenType() != lexer.LayerfileSKIP_REMAINING_IF_EOL; token = stream.Pop() {
		if token.GetTokenType() == lexer.LayerfileSKIP_REMAINING_IF_AND {
			if len(skipRemainingIf) == 0 {
				return nil, fmt.Errorf("AND must not be the first value of SKIP REMAINING IF")
			}
			if strings.TrimSpace(skipRemainingIf[len(skipRemainingIf)-1]) == "AND" {
				return nil, fmt.Errorf("AND must not be following AND in SKIP REMAINING IF")
			}
		}
		tokenValue := strings.TrimSpace(token.GetText())
		submatches := regexp.MustCompile("(.*?)(!=~|=~|!=|=)(.*)").FindStringSubmatch(tokenValue)
		//fmt.Println(tokenValue, ", ", len(submatches))
		if submatches != nil {
			for i, submatch := range submatches {
				submatch = strings.ReplaceAll(submatch, "'", "")
				submatch = strings.ReplaceAll(submatch, "\"", "")
				submatch = strings.TrimSpace(submatch)
				submatches[i] = submatch
			}
			skipRemainingIf = append(skipRemainingIf, strings.Join(submatches[1:], ""))
		} else { //"AND" or something like that
			skipRemainingIf = append(skipRemainingIf, tokenValue)
		}
	}
	if len(skipRemainingIf) == 0 {
		return nil, fmt.Errorf("SKIP REMAINING IF must be followed by at least one value")
	}

	return &SkipRemainingIf{SkipRemainingIf: skipRemainingIf}, nil
}

func (skipRemainingIf *SkipRemainingIf) String() string {
	var buf bytes.Buffer
	buf.WriteString("SKIP REMAINING IF")
	for _, val := range skipRemainingIf.SkipRemainingIf {
		buf.WriteRune(' ')
		if eqIdx := strings.Index(val, "="); eqIdx >= 0 {
			buf.WriteString(val[:eqIdx])
			buf.WriteRune('=')
			buf.WriteRune('"')
			if eqIdx < len(val)-1 {
				buf.WriteString(val[eqIdx+1:])
			}
			buf.WriteRune('"')
		} else {
			buf.WriteString(val)
		}
	}
	return buf.String()
}

func (skipRemainingIf *SkipRemainingIf) Hash(h hash.Hash, context *hashcontext.HashContext) {
	//do nothing - SKIP REMAINING IF do not edit the VM
}
