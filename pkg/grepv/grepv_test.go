package grepv_test

import (
	"bytes"
	"github.com/webappio/layerfiles/pkg/grepv"
	"testing"
)

func TestNoMatch(t *testing.T) {
	t.Parallel()

	buf := bytes.Buffer{}
	out := grepv.New([]byte("start"), []byte("end"), &buf)

	written := "some tokens now zzzzzzzz"
	_, err := out.Write([]byte(written))
	if err != nil {
		t.Error(err)
	}

	if buf.String() != "some tokens now zzzzzzzz" {
		t.Errorf("the output was not correctly written, was: %s, expected: %s", buf.String(), written)
	}
}

func TestSimpleMatch(t *testing.T) {
	t.Parallel()

	buf := bytes.Buffer{}
	out := grepv.New([]byte("start"), []byte("end"), &buf)

	written := "startaaaend"
	_, err := out.Write([]byte(written))
	if err != nil {
		t.Error(err)
	}

	if buf.String() != "" {
		t.Errorf("the output was not deleted, was: %s", buf.String())
	}
}

func TestEmptyMatch(t *testing.T) {
	t.Parallel()

	buf := bytes.Buffer{}
	out := grepv.New([]byte("s"), []byte("e"), &buf)

	written := "se"
	_, err := out.Write([]byte(written))
	if err != nil {
		t.Error(err)
	}

	if buf.String() != "" {
		t.Errorf("the output was not deleted, was: %s", buf.String())
	}
}

func TestExternalMatch(t *testing.T) {
	t.Parallel()

	buf := bytes.Buffer{}
	out := grepv.New([]byte("start"), []byte("end"), &buf)

	written := "aastartaaaendaa"
	_, err := out.Write([]byte(written))
	if err != nil {
		t.Error(err)
	}

	if buf.String() != "aaaa" {
		t.Errorf("the output was not deleted, was: '%s', should be 'aaaa'", buf.String())
	}
}

func TestEndBeforeStart(t *testing.T) {
	t.Parallel()

	buf := bytes.Buffer{}
	out := grepv.New([]byte("start"), []byte("end"), &buf)

	written := "endaaastartaa"
	_, err := out.Write([]byte(written))
	if err != nil {
		t.Error(err)
	}

	if buf.String() != "endaaa" {
		t.Errorf("the output was not deleted, was: '%s', should be 'endaaa'", buf.String())
	}
}

func TestSimpleMultiWrite(t *testing.T) {
	t.Parallel()

	buf := bytes.Buffer{}
	out := grepv.New([]byte("start"), []byte("end"), &buf)

	_, err := out.Write([]byte("sta"))
	if err != nil {
		t.Error(err)
	}
	_, err = out.Write([]byte("rt some data e"))
	if err != nil {
		t.Error(err)
	}
	_, err = out.Write([]byte("nd more data"))
	if err != nil {
		t.Error(err)
	}

	if buf.String() != " more data" {
		t.Errorf("the output was not correct, was: '%s', should be ' more data'", buf.String())
	}
}

func TestComplexMultiWrite(t *testing.T) {
	t.Parallel()

	buf := bytes.Buffer{}
	out := grepv.New([]byte("start"), []byte("end"), &buf)

	write := func(in string) {
		_, err := out.Write([]byte(in))
		if err != nil {
			t.Error(err)
		}
	}

	write("xxsta")
	write("rt some ststar star")
	write("t more data en")
	write("d sta end end start")
	write("even more data")
	write("end last data")

	if buf.String() != "xx sta end end  last data" {
		t.Errorf("the output was not correct, was: '%s', should be 'xx sta end end  last data'", buf.String())
	}
}
