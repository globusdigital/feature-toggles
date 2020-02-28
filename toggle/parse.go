package toggle

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"unicode"
)

type kind int

const (
	ident kind = iota // identifier

	intLit    // integer
	floatLit  // float
	boolLit   // boolean
	stringLit // string

	andOp // && operator
	orOp  // || operator
	eqOp  // == operator
	neOp  // != operator
	ltOp  // < operator
	gtOp  // > operator

	openParen  // (
	closeParen // )
)

type token struct {
	kind   kind
	pos    int
	opened rune
	val    []byte
}

func (t token) String() string {
	return fmt.Sprintf("type %q with value %q at position %d", t.kind, string(t.val), t.pos)
}

func ParseCondition(r io.Reader) (Condition, error) {
	tokens, err := lexer(r)
	if err != nil {
		return Condition{}, err
	}

	condition, pos, err := parseCondition(tokens, true, 0)
	if err != nil {
		return condition, err
	}

	if pos < len(tokens) {
		return condition, fmt.Errorf("trailing tokens: %s", tokens[pos:])
	}

	if condition.Op == invalidConditionalOp {
		condition.Op = AndOp
	}

	return condition, nil
}

func parseCondition(tokens []*token, toplevel bool, openP int) (Condition, int, error) {
	c := Condition{Op: invalidConditionalOp}
	i := 0

	for ; i < len(tokens); i++ {
		t := tokens[i]

		switch t.kind {
		case ident, eqOp, neOp, ltOp, gtOp, intLit, floatLit, boolLit, stringLit:
			f, pos, err := parseField(tokens[i:])
			if err != nil {
				return c, i, err
			}

			i += pos

			c.Fields = append(c.Fields, f)
		case andOp, orOp:
			op := AndOp
			if t.kind == orOp {
				op = OrOp
			}

			if c.Op == invalidConditionalOp || c.Op == op {
				c.Op = op
			} else {
				if OrOp == op && !toplevel {
					return c, i, nil
				}

				condition, pos, err := parseCondition(tokens[i+1:], false, openP)
				if err != nil {
					return c, i, err
				}

				if t.kind == andOp {
					if len(c.Fields) > 0 {
						// |-------------------------| parent OR condition
						// foo == true || bar == false && alpha < 42
						//               ^------------^ becomes part of the AND condition
						condition.Fields = append(
							[]ConditionField{c.Fields[len(c.Fields)-1]}, condition.Fields...)
						if len(c.Fields) == 1 {
							c.Fields = nil
						} else {
							c.Fields = c.Fields[:len(c.Fields)-1]

						}

						condition.Op = AndOp
					}

					if condition.Op != invalidConditionalOp {
						c.Conditions = append(c.Conditions, condition)
					}
				} else {
					if condition.Op == invalidConditionalOp {
						// Trailing field
						c = Condition{Op: op, Conditions: []Condition{c}, Fields: condition.Fields}
					} else {
						c = Condition{Op: op, Conditions: []Condition{c, condition}}
					}
				}

				i += pos + 1
			}
		case openParen:
			condition, pos, err := parseCondition(tokens[i+1:], true, openP+1)
			if err != nil {
				return c, i, err
			}

			c.Conditions = append(c.Conditions, condition)
			i += pos + 1

		case closeParen:
			if openP < 1 {
				return c, i, fmt.Errorf("unexpected token (%s)", t)
			}
			return c, i, nil
		}
	}

	if openP > 0 {
		return c, i, fmt.Errorf("imbalanced opening parenthesis")
	}

	return c, i, nil
}

func parseField(tokens []*token) (ConditionField, int, error) {
	f := ConditionField{Op: invalidFieldOp}
	i := 0

LOOP:
	for ; i < len(tokens); i++ {
		t := tokens[i]

		switch t.kind {
		default:
			i--
			break LOOP
		case ident:
			// name .
			if f.Name != "" ||
				// value .
				f.Name == "" && f.Value != nil && f.Op == invalidFieldOp {
				return f, i, fmt.Errorf("unexpected token (%s)", t)
			}
			f.Name = string(t.val)
		case eqOp, neOp, ltOp, gtOp:
			// op .
			if f.Op != invalidFieldOp ||
				// .
				f.Name == "" && f.Value == nil ||
				// name value .
				f.Name != "" && f.Value != nil {
				return f, i, fmt.Errorf("unexpected token (%s)", t)
			}

			switch t.kind {
			case eqOp:
				f.Op = EqOp
			case neOp:
				f.Op = NeOp
			case ltOp:
				// Value is the first token
				if f.Name == "" {
					f.Op = GtOp
				} else {
					f.Op = LtOp
				}
			case gtOp:
				// Value is the first token
				if f.Name == "" {
					f.Op = LtOp
				} else {
					f.Op = GtOp
				}
			}
		case intLit, floatLit, boolLit, stringLit:
			// value .
			if f.Value != nil ||
				// name .
				f.Value == nil && f.Name != "" && f.Op == invalidFieldOp {
				return f, i, fmt.Errorf("unexpected token (%s)", t)
			}

			var err error
			switch t.kind {
			case stringLit:
				f.Value = string(t.val)
				f.Type = StringType
			case intLit:
				f.Value, err = strconv.ParseInt(string(t.val), 10, 64)
				f.Type = IntType
			case floatLit:
				f.Value, err = strconv.ParseFloat(string(t.val), 10)
				f.Type = FloatType
			case boolLit:
				f.Type = BoolType
				if bytes.Equal(t.val, []byte("true")) {
					f.Value = true
				} else if bytes.Equal(t.val, []byte("false")) {
					f.Value = false
				} else {
					err = fmt.Errorf("invalid boolean value: %s", string(t.val))
				}
			}

			if err != nil {
				return f, i, fmt.Errorf("unexpected value for token (%s): %v", t, err)
			}
		}
	}

	if f.Name == "" || f.Op == invalidFieldOp || f.Value == nil {
		return f, i, fmt.Errorf("incomplete field: %s", f)
	}

	return f, i, nil
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
					t.val, escapeNext = addRuneToString(t.val, r, escapeNext)
				} else {
					t = nil
				}
			}
		case r == '(':
			if t == nil {
				t = &token{kind: openParen, pos: curLen}
			} else if t.kind == stringLit {
				t.val, escapeNext = addRuneToString(t.val, r, escapeNext)
			} else {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
		case r == ')':
			if t == nil {
				t = &token{kind: closeParen, pos: curLen}
			} else if t.kind == stringLit {
				t.val, escapeNext = addRuneToString(t.val, r, escapeNext)
			} else {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
		case r == '<':
			if t == nil {
				t = &token{kind: ltOp, pos: curLen}
			} else if t.kind == stringLit {
				t.val, escapeNext = addRuneToString(t.val, r, escapeNext)
			} else {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
		case r == '>':
			if t == nil {
				t = &token{kind: gtOp, pos: curLen}
			} else if t.kind == stringLit {
				t.val, escapeNext = addRuneToString(t.val, r, escapeNext)
			} else {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
		case r == '&':
			if t == nil {
				t = &token{kind: andOp, pos: curLen, val: []byte(string(r))}
			} else if t.kind == stringLit {
				t.val, escapeNext = addRuneToString(t.val, r, escapeNext)
			} else if t.kind != andOp || len(t.val) != 1 {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			} else {
				t.val = append(t.val, []byte(string(r))...)
			}
		case r == '|':
			if t == nil {
				t = &token{kind: orOp, pos: curLen, val: []byte(string(r))}
			} else if t.kind == stringLit {
				t.val, escapeNext = addRuneToString(t.val, r, escapeNext)
			} else if t.kind != orOp || len(t.val) != 1 {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			} else {
				t.val = append(t.val, []byte(string(r))...)
			}
		case r == '!':
			if t == nil {
				t = &token{kind: neOp, pos: curLen, val: []byte(string(r))}
			} else if t.kind == stringLit {
				t.val, escapeNext = addRuneToString(t.val, r, escapeNext)
			} else {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
		case r == '=':
			if t == nil {
				t = &token{kind: eqOp, pos: curLen, val: []byte(string(r))}
			} else if t.kind == stringLit {
				t.val, escapeNext = addRuneToString(t.val, r, escapeNext)
			} else if (t.kind != eqOp && t.kind != neOp) || len(t.val) != 1 {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			} else {
				t.val = append(t.val, []byte(string(r))...)
			}
		case r == '\\':
			if t == nil || t.kind != stringLit {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
			if escapeNext {
				t.val = append(t.val, '\\', '\\')
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
					if t.opened != r && escapeNext {
						t.val = append(t.val, '\\')
					}
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
				t.val, escapeNext = addRuneToString(t.val, r, escapeNext)
			case ident:
				t.val = append(t.val, []byte(string(r))...)
			default:
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
		case unicode.IsNumber(r):
			if t == nil {
				t = &token{kind: intLit, pos: curLen}
			}

			switch t.kind {
			case stringLit:
				t.val, escapeNext = addRuneToString(t.val, r, escapeNext)
			case ident, intLit, floatLit:
				t.val = append(t.val, []byte(string(r))...)
			default:
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
		case r == '.':
			if t == nil {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}

			switch t.kind {
			case intLit:
				t.kind = floatLit
				t.val = append(t.val, []byte(string(r))...)
			case stringLit:
				t.val, escapeNext = addRuneToString(t.val, r, escapeNext)
			default:
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
		default:
			if t == nil || t.kind != stringLit {
				return tokens, fmt.Errorf("invalid character %v at position %d", string(r), curLen)
			}
			t.val, escapeNext = addRuneToString(t.val, r, escapeNext)
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

func addRuneToString(buf []byte, r rune, escapeNext bool) ([]byte, bool) {
	if escapeNext {
		buf = append(buf, '\\')
		escapeNext = false
	}
	buf = append(buf, []byte(string(r))...)

	return buf, escapeNext
}
