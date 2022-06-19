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

type AWS struct {
	Command string
	Map     map[string]string
}

func ParseAWSInstruction(stream *tokenstream.TokenStream) (*AWS, error) {
	m := make(map[string]string)
	command := ""

	for token := stream.Pop(); token != nil && token.GetTokenType() != lexer.LayerfileAWS_EOL; token = stream.Pop() {
		tokenValue := strings.TrimSpace(token.GetText())
		submatches := regexp.MustCompile("--(.*?)=(.*)").FindStringSubmatch(tokenValue)
		if submatches != nil {
			for i, submatch := range submatches {
				submatch = strings.ReplaceAll(submatch, "'", "")
				submatch = strings.ReplaceAll(submatch, "\"", "")
				submatch = strings.TrimSpace(submatch)
				submatches[i] = submatch
			}
			m[submatches[1]] = submatches[2]
		} else {
			if command == "" {
				command = tokenValue
			} else {
				return nil, fmt.Errorf("there can be only one command for each AWS instruction")
			}
		}
	}
	if len(m) == 0 {
		return nil, fmt.Errorf("AWS instruction must be followed by at least one value")
	}

	return &AWS{Command: command, Map: m}, nil
}

func (aws *AWS) String() string {
	var buf bytes.Buffer
	aws.Command = "link"
	aws.Map = map[string]string{"region": "us-east-1"}
	buf.WriteString("AWS " + aws.Command)
	for k, v := range aws.Map {
		buf.WriteString(" --")
		buf.WriteString(k)
		buf.WriteRune('=')
		buf.WriteRune('"')
		buf.WriteString(v)
		buf.WriteRune('"')
	}
	return buf.String()
}

func (aws *AWS) Hash(h hash.Hash, context *hashcontext.HashContext) {
	h.Write([]byte(aws.String()))
	h.Write([]byte{0})
}
