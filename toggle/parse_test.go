package toggle

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const cond1 = `userID < 10 && (serviceName == 'serv1' || serviceName == 'serv2') || userGroup == "tes\"t'er"`

func Test_lexer(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []*token
		wantErr bool
	}{
		{name: "ident", in: "userID", want: []*token{{kind: ident, pos: 0, val: []byte("userID")}}},
		{name: "intLit", in: "10", want: []*token{{kind: intLit, pos: 0, val: []byte("10")}}},
		{name: "floatLit", in: "10.1", want: []*token{{kind: floatLit, pos: 0, val: []byte("10.1")}}},
		{name: "boolLit", in: "true", want: []*token{{kind: boolLit, pos: 0, val: []byte("true")}}},
		{name: "boolLit", in: "false", want: []*token{{kind: boolLit, pos: 0, val: []byte("false")}}},
		{name: "stringLit", in: "'false'", want: []*token{{kind: stringLit, pos: 0, val: []byte("false"), opened: '\''}}},
		{name: "stringLit 2", in: `"true != \"false\""`, want: []*token{{kind: stringLit, pos: 0, val: []byte("true != \"false\""), opened: '"'}}},
		{name: "cond1", in: cond1, want: []*token{
			{kind: ident, pos: 0, val: []byte("userID")},
			{kind: ltOp, pos: 7},
			{kind: intLit, pos: 9, val: []byte("10")},
			{kind: andOp, pos: 12, val: []byte("&&")},
			{kind: openParen, pos: 15},
			{kind: ident, pos: 16, val: []byte("serviceName")},
			{kind: eqOp, pos: 28, val: []byte("==")},
			{kind: stringLit, pos: 31, val: []byte("serv1"), opened: '\''},
			{kind: orOp, pos: 39, val: []byte("||")},
			{kind: ident, pos: 42, val: []byte("serviceName")},
			{kind: eqOp, pos: 54, val: []byte("==")},
			{kind: stringLit, pos: 57, val: []byte("serv2"), opened: '\''},
			{kind: closeParen, pos: 64},
			{kind: orOp, pos: 66, val: []byte("||")},
			{kind: ident, pos: 69, val: []byte("userGroup")},
			{kind: eqOp, pos: 79, val: []byte("==")},
			{kind: stringLit, pos: 82, val: []byte(`tes"t'er`), opened: '"'},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lexer(strings.NewReader(tt.in))
			if (err != nil) != tt.wantErr {
				t.Errorf("lexer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
