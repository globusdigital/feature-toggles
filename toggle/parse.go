package toggle

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode"
)

type kind int

const (
	ident kind = iota

	intLit
	floatLit
	boolLit
	stringLit

	andOp
	orOp
	eqOp
	neOp
	ltOp
	gtOp

	openParen
	closeParen
)

type token struct {
	kind   kind
	pos    int
	opened rune
	val    []byte
}

func lexer(r io.Reader) ([]*token, error) {
	br := bufio.NewReader(r)

	var tokens []*token
	var t *token
	var curLen int
	var escapeNext bool
	for {
		r, l, err := br.ReadRune()
		if err != nil {
			if err != io.EOF {
				return tokens, err
			}
			break
		}

		switch {
		case unicode.IsSpace(r):
			if t != nil {
				if t.kind == ident {
					if bytes.Equal(t.val, []byte("true")) || bytes.Equal(t.val, []byte("false")) {
						t.kind = boolLit
					}
				}

				if t.kind == stringLit {
					t.val = append(t.val, []byte(string(r))...)
				} else {
					t = nil
				}
			}
		case r == '(':
			if t == nil {
				t = &token{kind: openParen, pos: curLen}
			} else if t.kind == stringLit {
				t.val = append(t.val, []byte(string(r))...)
			} else {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
		case r == ')':
			if t == nil {
				t = &token{kind: closeParen, pos: curLen}
			} else if t.kind == stringLit {
				t.val = append(t.val, []byte(string(r))...)
			} else {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
		case r == '<':
			if t == nil {
				t = &token{kind: ltOp, pos: curLen}
			} else if t.kind == stringLit {
				t.val = append(t.val, []byte(string(r))...)
			} else {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
		case r == '>':
			if t == nil {
				t = &token{kind: gtOp, pos: curLen}
			} else if t.kind == stringLit {
				t.val = append(t.val, []byte(string(r))...)
			} else {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
		case r == '&':
			if t == nil {
				t = &token{kind: andOp, pos: curLen}
			} else if t.kind != stringLit && (t.kind != andOp || len(t.val) != 1) {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
			t.val = append(t.val, []byte(string(r))...)
		case r == '|':
			if t == nil {
				t = &token{kind: orOp, pos: curLen}
			} else if t.kind != stringLit && (t.kind != orOp || len(t.val) != 1) {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
			t.val = append(t.val, []byte(string(r))...)
		case r == '!':
			if t == nil {
				t = &token{kind: neOp, pos: curLen}
			} else if t.kind != stringLit {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
			t.val = append(t.val, []byte(string(r))...)
		case r == '=':
			if t == nil {
				t = &token{kind: eqOp, pos: curLen}
			} else if t.kind != stringLit && !((t.kind == eqOp || t.kind == neOp) && len(t.val) == 1) {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
			t.val = append(t.val, []byte(string(r))...)
		case r == '\\':
			if t == nil || t.kind != stringLit {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
			if escapeNext {
				t.val = append(t.val, []byte(string(r))...)
				escapeNext = false
			} else {
				escapeNext = true
			}
		case r == '\'' || r == '"':
			if t == nil {
				t = &token{kind: stringLit, pos: curLen, opened: r}
			} else if t.kind != stringLit {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}

			if len(t.val) > 0 {
				if escapeNext || t.opened != r {
					t.val = append(t.val, []byte(string(r))...)
					escapeNext = false
				} else {
					t = nil
				}
			}
		case unicode.IsLetter(r) || r == '_':
			if t == nil {
				t = &token{kind: ident, pos: curLen}
			}
			switch t.kind {
			case stringLit:
				if escapeNext {
					t.val = append(t.val, '\\')
					escapeNext = false
				}
			case ident:
			default:
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
			t.val = append(t.val, []byte(string(r))...)
		case unicode.IsNumber(r):
			if t == nil {
				t = &token{kind: intLit, pos: curLen}
			}

			switch t.kind {
			case stringLit:
				if escapeNext {
					t.val = append(t.val, '\\')
					escapeNext = false
				}
			case ident, intLit, floatLit:
			default:
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
			t.val = append(t.val, []byte(string(r))...)
		case r == '.':
			if t == nil {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}

			switch t.kind {
			case intLit:
				t.kind = floatLit
			case stringLit:
				if escapeNext {
					t.val = append(t.val, '\\')
					escapeNext = false
				}
			default:
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
			t.val = append(t.val, []byte(string(r))...)
		default:
			if t == nil || t.kind != stringLit {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
			if escapeNext {
				t.val = append(t.val, '\\')
				escapeNext = false
			}
			t.val = append(t.val, []byte(string(r))...)
		}

		if t != nil && (len(tokens) == 0 || t != tokens[len(tokens)-1]) {
			tokens = append(tokens, t)
		}
		if t != nil {
			switch t.kind {
			case openParen, closeParen, ltOp, gtOp:
				t = nil
			case andOp, orOp, eqOp, neOp:
				if len(t.val) == 2 {
					t = nil
				}
			}
		}
		curLen += l
	}
	if t != nil && t.kind == ident {
		if bytes.Equal(t.val, []byte("true")) || bytes.Equal(t.val, []byte("false")) {
			t.kind = boolLit
		}
	}

	return tokens, nil
}
