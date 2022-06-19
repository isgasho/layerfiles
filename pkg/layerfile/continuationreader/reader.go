package continuationreader

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"io"
	"io/ioutil"
	"regexp"
)

//New returns a new ContinuationReader, which allows escaping newlines
//======= INPUT =========
//FROM hello\
//world:\
//latest
//
//RUN echo "this is\ some\ \
//command"
//
//======= OUTPUT =========
//FROM helloworld:latest
//
//RUN echo "this is some command"

var readerRegex = regexp.MustCompile(`\\[ \t]*(\r\n|\n|\r|$)`)

func New(r io.Reader) (*antlr.InputStream, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error while copying: %s", err.Error())
	}

	res := readerRegex.ReplaceAll(buf, []byte{})

	return antlr.NewInputStream(string(res)), nil
}
