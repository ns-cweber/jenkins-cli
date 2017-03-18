package boolsearch

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type SyntaxError int

func (err SyntaxError) Error() string {
	return fmt.Sprintf("Syntax error at %d", err)
}

type tokenType int

const (
	ttIdent tokenType = iota
	ttAnd
	ttOr
	ttNe
	ttEq
)

type Token struct {
	Typ tokenType
	Str string
}

var tokEq = Token{ttEq, "="}
var tokNe = Token{ttNe, "!="}
var tokOr = Token{ttOr, "|"}
var tokAnd = Token{ttAnd, "&"}

func Tokenize(r *bufio.Reader) ([]Token, error) {
	var tokens []Token
	var tok []rune
	var cursor int

NEXT_CHAR:
	for {
		ch, sz, err := r.ReadRune()
		if err != nil {
			if err == io.EOF {
				if len(tok) > 0 {
					tokens = append(tokens, Token{ttIdent, string(tok)})
					tok = tok[0:0]
				}
				break
			}
			return tokens, err
		}
		cursor += sz

		// If it's a space, append any in-progress token and move on
		if unicode.IsSpace(ch) {
			if len(tok) > 0 {
				tokens = append(tokens, Token{ttIdent, string(tok)})
				tok = tok[0:0]
			}
			continue NEXT_CHAR
		}

		// If it's a single-char token, append any in-progress token as well
		// as the new single-char token and move on
		if ch == '&' {
			if len(tok) > 0 {
				tokens = append(tokens, Token{ttIdent, string(tok)})
				tok = tok[0:0]
			}
			tokens = append(tokens, tokAnd)
			continue NEXT_CHAR
		}
		if ch == '|' {
			if len(tok) > 0 {
				tokens = append(tokens, Token{ttIdent, string(tok)})
				tok = tok[0:0]
			}
			tokens = append(tokens, tokOr)
			continue NEXT_CHAR
		}
		if ch == '=' {
			if len(tok) > 0 {
				tokens = append(tokens, Token{ttIdent, string(tok)})
				tok = tok[0:0]
			}
			tokens = append(tokens, tokEq)
			continue NEXT_CHAR
		}

		// If the character is a letter or a number, add it to the token and
		// move on
		if unicode.IsLetter(ch) || unicode.IsNumber(ch) {
			tok = append(tok, ch)
			continue NEXT_CHAR
		}

		// Handle '!='
		if ch == '!' {
			if len(tok) > 0 {
				tokens = append(tokens, Token{ttIdent, string(tok)})
				tok = tok[0:0]
			}
			ch, sz, err := r.ReadRune()
			if err != nil {
				return tokens, err
			}
			cursor += sz
			if ch == '=' {
				tokens = append(tokens, tokNe)
				continue NEXT_CHAR
			}
		}

		return tokens, SyntaxError(cursor)
	}
	return tokens, nil
}

func parseComparison(toks []Token) (Expression, error) {
	if toks[0].Typ != ttIdent || toks[2].Typ != ttIdent {
		return nil, fmt.Errorf("Invalid comparison expression: %v", toks)
	}

	var op CmpOp
	switch toks[1].Typ {
	case ttEq:
		op = CmpOpEq
	case ttNe:
		op = CmpOpNe
	default:
		return nil, fmt.Errorf("Invalid comparison expression: %v", toks)
	}

	return Comparison{Left: toks[0].Str, Right: toks[2].Str, Op: op}, nil
}

func parseExpr(toks []Token) (Expression, error) {
	if len(toks) < 3 {
		return nil, fmt.Errorf("Invalid expression: %v", toks)
	}

	expr, err := parseComparison(toks[:3])
	if err != nil {
		return nil, err
	}

	toks = toks[3:]

	if len(toks) < 1 {
		return expr, nil
	}

	var op ConjOp
	switch toks[0].Typ {
	case ttAnd:
		op = ConjOpAnd
	case ttOr:
		op = ConjOpOr
	default:
		return nil, fmt.Errorf("Invalid conjugation operator: %s", toks[0].Str)
	}

	right, err := parseExpr(toks[1:])
	if err != nil {
		return nil, err
	}

	return Conjugation{Op: op, Left: expr, Right: right}, nil
}

func ParseTokens(toks []Token) (Expression, error) {
	if len(toks) < 1 {
		return Empty{}, nil
	}

	return parseExpr(toks)
}

func ParseReader(r io.Reader) (Expression, error) {
	tokens, err := Tokenize(bufio.NewReader(r))
	if err != nil {
		return nil, err
	}
	return ParseTokens(tokens)
}

func ParseBytes(p []byte) (Expression, error) {
	return ParseReader(bytes.NewReader(p))
}

func ParseString(s string) (Expression, error) {
	return ParseReader(strings.NewReader(s))
}
