package layerfile

import (
	"bytes"
	"crypto/sha256"
	"github.com/webappio/layerfiles/pkg/layerfile/hashcontext"
	"io/ioutil"
	"os"
	"testing"
)

const SanicLayerfile = `
FROM ubuntu:18.04

RUN apt-get update
RUN apt-get install -y ca-certificates curl
CHECKPOINT

# install go
RUN curl -L "https://golang.org/dl/go1.14.6.linux-amd64.tar.gz" |\
    tar -C /usr/local -xzf /dev/stdin
CHECKPOINT

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH
ENV GO111MODULE=on

WORKDIR /app
COPY pkg main.go go.mod go.sum build.sh ./
RUN bash ./build.sh
CHECKPOINT

COPY example /example
WORKDIR /example/timestamp-as-a-service
RUN sanic env dev
RUN sanic run print_dev | grep "in dev!"
CHECKPOINT

# test whitespace doesn't break things:'
# tab here
	
# some spaces
    

`

func TestSanicLayerfile(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "layerci_test")
	if err != nil {
		t.Error(err)
	}
	defer func() { _ = os.Remove(tmpdir) }()

	err = os.MkdirAll(tmpdir+"/example/one", 0755)
	if err != nil {
		t.Error(err)
	}

	err = os.MkdirAll(tmpdir+"/pkg", 0755)
	if err != nil {
		t.Error(err)
	}

	tryMakeFile := func(filename string) {
		f, err := os.OpenFile(tmpdir+"/"+filename, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			t.Error(err)
		}
		_, err = f.WriteString(filename)
		if err != nil {
			t.Error(err)
		}
		err = f.Close()
		if err != nil {
			t.Error(err)
		}
	}

	tryMakeFile("main.go")
	tryMakeFile("go.mod")
	tryMakeFile("go.sum")
	tryMakeFile("build.sh")
	tryMakeFile("pkg/filea")
	tryMakeFile("pkg/fileb")
	tryMakeFile("example/one/somefile")

	tokens, err := tokenizeLayerfile(bytes.NewBufferString(SanicLayerfile))
	if err != nil {
		t.Error(err)
		return
	}

	instrs, err := parseInstructions(tokens)
	if err != nil {
		t.Error(err)
	}

	_, instrs, err = parseFrom(instrs)
	if err != nil {
		t.Error(err)
	}

	hasher := sha256.New()
	for _, instr := range instrs {
		instr.Hash(hasher, &hashcontext.HashContext{})
	}
}
