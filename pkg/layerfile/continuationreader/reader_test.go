package continuationreader

import (
	"bytes"
	"testing"
)

func TestReaderLinux(t *testing.T) {

	input := `
some\
text  with\ \ some \
contin\
\
uations and sometimes no
\continuations

/bin/sh build/lib/test.sh \
    & docker build --build-arg VERSION=${JOB_ID} \
    -t roostliving/puppeteer:${GIT_BRANCH//\//-} \
    -t roostliving/puppeteer:${JOB_ID} \
    -f build/puppeteer/Dockerfile . \
    & docker build --network ${JOB_ID}_default \ 
    -t roostliving/web:${GIT_BRANCH//\//-} \
    -t roostliving/web:${JOB_ID} \
    -f build/web/Dockerfile .
`

	expected := `
sometext  with\ \ some continuations and sometimes no
\continuations

/bin/sh build/lib/test.sh     & docker build --build-arg VERSION=${JOB_ID}     -t roostliving/puppeteer:${GIT_BRANCH//\//-}     -t roostliving/puppeteer:${JOB_ID}     -f build/puppeteer/Dockerfile .     & docker build --network ${JOB_ID}_default     -t roostliving/web:${GIT_BRANCH//\//-}     -t roostliving/web:${JOB_ID}     -f build/web/Dockerfile .
`

	reader, err := New(bytes.NewBufferString(input))
	if err != nil {
		t.Error(err)
		return
	}
	if reader.String() != expected {
		t.Fatalf("Got improperly continued string: \n%#v\n - expected: \n%#v", reader.String(), expected)
	}
}

func TestReaderOtherLineEndings(t *testing.T) {
	input := "a\\\r\nb\\\rc\\\n\r\\d\\\n\n"
	expected := "abc\r\\d\n"

	reader, err := New(bytes.NewBufferString(input))
	if err != nil {
		t.Error(err)
		return
	}
	if reader.String() != expected {
		t.Fatalf("Got improperly continued string: %q", reader.String())
	}
}
