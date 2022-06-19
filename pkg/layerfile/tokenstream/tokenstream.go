package tokenstream

import (
	"bytes"
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type TokenStream struct {
	tokens []antlr.Token
	idx    int
}

func Read(stream *antlr.CommonTokenStream) (*TokenStream, error) {
	var tokens []antlr.Token
	for _, token := range stream.GetAllTokens() {
		if token.GetTokenType() >= antlr.TokenMinUserTokenType {
			tokens = append(tokens, token)
		} else if token.GetTokenType() == antlr.TokenInvalidType {
			return nil, fmt.Errorf("error with '%s' at line %d:%d", token.GetText(), token.GetLine(), token.GetColumn())
		}
	}
	return &TokenStream{tokens: tokens}, nil
}

func (s *TokenStream) Peek() antlr.Token {
	if s.idx < 0 || s.idx >= len(s.tokens) {
		return nil
	}
	return s.tokens[s.idx]
}

func (s *TokenStream) HasToken() bool {
	return s.idx < len(s.tokens)
}

func (s *TokenStream) Pop() antlr.Token {
	token := s.Peek()
	s.idx += 1
	return token
}

func (s *TokenStream) String() string {
	var buf bytes.Buffer
	for _, token := range s.tokens {
		buf.WriteString(token.GetText())
		buf.WriteString(" ")
	}
	return buf.String()
}
