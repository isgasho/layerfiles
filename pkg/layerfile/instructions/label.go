package instructions

import (
	"bytes"
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
	"strings"
)

type Label struct {
	Label []string
}

func CheckAllowableSyntax(key string, val string) error {
	key = strings.TrimSpace(strings.ToLower(key))
	val = strings.TrimSpace(strings.ToLower(val))
	switch key {
	case "status":
		if val != "merge" && val != "hidden" {
			return fmt.Errorf("LABEL status= must contain merge or hidden (UNKNOWN VALUE: %s)", val)
		}
		break
	case "display_name":
		break
	default:
		return fmt.Errorf("LABEL unrecognized key %s", key)
	}
	return nil
}

func ParseLabelInstruction(stream *tokenstream.TokenStream) (*Label, error) {
	label := []string{}
	prevToken := ""
	for token := stream.Pop(); token != nil && token.GetTokenType() != lexer.LayerfileLABEL_EOL; token = stream.Pop() {
		tokenValue := token.GetText()
		if strings.Contains(tokenValue, "=") {
			label = append(label, tokenValue)
			keyVal := strings.Split(tokenValue, "=")
			if err := CheckAllowableSyntax(keyVal[0], keyVal[1]); err != nil {
				return nil, err
			}
		} else if prevToken != "" {
			label = append(label, prevToken+"="+tokenValue)
			prevToken = ""
			if err := CheckAllowableSyntax(prevToken, tokenValue); err != nil {
				return nil, err
			}
		} else {
			prevToken = tokenValue
		}
	}
	if len(label) == 0 {
		return nil, fmt.Errorf("LABEL must be followed by at least one var=value pair")
	}
	return &Label{Label: label}, nil
}

func (label *Label) String() string {
	var buf bytes.Buffer
	buf.WriteString("LABEL")
	for _, val := range label.Label {
		buf.WriteRune(' ')
		buf.WriteString(val)
	}
	return buf.String()
}

func (label *Label) Hash(h hash.Hash, context *hashcontext.HashContext) {
}
