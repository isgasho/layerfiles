package instructions

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"github.com/webappio/layerfiles/pkg/layerfile/tokenstream"
	"hash"
	"regexp"
	"strconv"
	"strings"
)

type ExposeWebsite struct {
	Scheme string
	Domain string
	Port   uint16

	Path        string
	RewritePath string
}

var websiteAddrRegex = regexp.MustCompile("^(?:(https?)://)?([^ \t:/]+)(?::(\\d+))?$")

func ParseExposeWebsiteInstruction(stream *tokenstream.TokenStream) (*ExposeWebsite, error) {
	token := stream.Pop()
	if token == nil {
		return nil, fmt.Errorf("EXPOSE WEBSITE requires a URL immediately after it")
	}
	if token.GetTokenType() != lexer.LayerfileWEBSITE_ITEM {
		return nil, fmt.Errorf("unexpected token while reading 'EXPOSE WEBSITE': %s", token.GetText())
	}

	match := websiteAddrRegex.FindStringSubmatch(token.GetText())
	if match == nil {
		return nil, fmt.Errorf("EXPOSE WEBSITE was of incorrect form: %v", token.GetText())
	}

	instr := &ExposeWebsite{
		Path:   "/",
		Domain: "localhost",
	}
	if match[2] != "localhost" {
		if port, err := strconv.ParseUint(match[2], 10, 16); err == nil {
			instr.Port = uint16(port)
		} else {
			return nil, fmt.Errorf("EXPOSE WEBSITE must expose 'localhost', not '%v'. Try EXPOSE WEBSITE 80", match[2])
		}
	}

	if match[1] == "" {
		instr.Scheme = "http"
	} else if match[1] != "http" && match[1] != "https" {
		return nil, fmt.Errorf("EXPOSE WEBSITE must be of the form http://localhost, not %v://localhost", match[0])
	} else {
		instr.Scheme = match[1]
	}

	if instr.Port == 0 {
		if match[3] == "" {
			if instr.Scheme == "http" {
				instr.Port = 80
			} else {
				instr.Port = 443
			}
		} else {
			parsed, err := strconv.ParseUint(match[3], 10, 16)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid port specified: %v", match[3])
			}
			instr.Port = uint16(parsed)
		}
	}

	token = stream.Pop()
	if token == nil || token.GetTokenType() == lexer.LayerfileWEBSITE_EOL {
		return instr, nil
	} else if token.GetTokenType() != lexer.LayerfileWEBSITE_ITEM {
		return nil, fmt.Errorf("invalid data in EXPOSE WEBSITE instruction: %v", token.GetText())
	}

	instr.Path = "/" + strings.TrimSpace(strings.Trim(token.GetText(), "/"))

	token = stream.Pop()
	if token == nil || token.GetTokenType() == lexer.LayerfileWEBSITE_EOL {
		return instr, nil
	} else if token.GetTokenType() != lexer.LayerfileWEBSITE_ITEM {
		return nil, fmt.Errorf("invalid data in EXPOSE WEBSITE instruction: %v", token.GetText())
	}

	instr.RewritePath = "/" + strings.TrimSpace(strings.TrimLeft(token.GetText(), "/"))

	token = stream.Pop()
	if token == nil || token.GetTokenType() == lexer.LayerfileWEBSITE_EOL {
		return instr, nil
	} else {
		return nil, fmt.Errorf("invalid data in EXPOSE WEBSITE instruction: %v", token.GetText())
	}
}

func (expose *ExposeWebsite) String() string {
	rewriteDest := ""
	if expose.RewritePath != "" {
		rewriteDest = " " + expose.RewritePath
	}
	return fmt.Sprintf("EXPOSE WEBSITE %s://%s:%d %s%s", expose.Scheme, expose.Domain, expose.Port, expose.Path, rewriteDest)
}

func (expose *ExposeWebsite) Hash(h hash.Hash, context *hashcontext.HashContext) {
	h.Write([]byte{0})
	h.Write([]byte(expose.String()))
}
