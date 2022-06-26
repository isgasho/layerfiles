package grepv

import (
	"bytes"
	"io"
)

//Grepv removes chunks of a stream between two given tokens
//(it buffers the input until it can be sure there is no match)
type Grepv struct {
	removeTokenStart []byte
	removeTokenEnd   []byte //i.e., []byte{'\n'}
	upstream         io.Writer

	waitingForTokenEnd bool

	retainedBuf bytes.Buffer
}

func New(start, end []byte, upstream io.Writer) *Grepv {
	return &Grepv{
		removeTokenStart: start,
		removeTokenEnd:   end,
		upstream:         upstream,
	}
}

func (grepv *Grepv) processTokenEnd() error {
	if !grepv.waitingForTokenEnd {
		return nil
	}
	idx := bytes.Index(grepv.retainedBuf.Bytes(), grepv.removeTokenEnd)
	if idx == -1 {
		return nil
	}
	grepv.retainedBuf.Next(idx + len(grepv.removeTokenEnd)) //discard up to the end of the token
	grepv.waitingForTokenEnd = false
	return grepv.processTokenStart()
}

func (grepv *Grepv) processTokenStart() error {
	if grepv.waitingForTokenEnd {
		return nil
	}
	idx := bytes.Index(grepv.retainedBuf.Bytes(), grepv.removeTokenStart)
	if idx == -1 {
		return nil
	}

	_, err := grepv.upstream.Write(grepv.retainedBuf.Next(idx)) //TODO no "n" handling
	grepv.retainedBuf.Next(len(grepv.removeTokenStart))         //discard the start token
	grepv.waitingForTokenEnd = true

	if err != nil {
		return err
	}

	return grepv.processTokenEnd()
}

func (grepv *Grepv) longestPotentialMatch() int {
	loopStart := len(grepv.removeTokenStart)
	if loopStart >= grepv.retainedBuf.Len() {
		loopStart = grepv.retainedBuf.Len()
	}
	for i := loopStart; i >= 1; i-- {
		if bytes.Equal(grepv.removeTokenStart[:i], grepv.retainedBuf.Bytes()[grepv.retainedBuf.Len()-i:]) {
			return i
		}
	}
	return 0
}

func (grepv *Grepv) Write(buf []byte) (int, error) {
	grepv.retainedBuf.Write(buf)

	err := grepv.processTokenEnd()
	if err != nil {
		return 0, err
	}

	err = grepv.processTokenStart()
	if err != nil {
		return 0, err
	}

	matchLength := grepv.longestPotentialMatch()
	if !grepv.waitingForTokenEnd {
		return grepv.upstream.Write(grepv.retainedBuf.Next(grepv.retainedBuf.Len() - matchLength))
	}
	return 0, err
}

func (grepv *Grepv) Close() error {
	_, _ = grepv.upstream.Write(grepv.retainedBuf.Next(grepv.retainedBuf.Len()))
	if closer, ok := grepv.upstream.(io.WriteCloser); ok {
		return closer.Close()
	}
	return nil
}
