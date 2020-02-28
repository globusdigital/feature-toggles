package toggle

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	cond1 = `userID < 10 && (serviceName == 'serv1' || serviceName == 'serv2') || userGroup == "tes\"t'er"`
	cond2 = `useriD < 10 && s == true || foo == "bar" && alpha == 14 || test == 42.1 && test2 == false`
)

var (
	cond1Exp = Condition{Op: OrOp, Conditions: []Condition{
		{Op: AndOp,
			Conditions: []Condition{
				{Op: OrOp, Fields: []ConditionField{
					{Op: EqOp, ConditionValue: ConditionValue{Name: "serviceName", Type: StringType, Value: "serv1"}},
					{Op: EqOp, ConditionValue: ConditionValue{Name: "serviceName", Type: StringType, Value: "serv2"}},
				}},
			},
			Fields: []ConditionField{
				{Op: LtOp, ConditionValue: ConditionValue{Name: "userID", Type: IntType, Value: int64(10)}},
			},
		},
	}, Fields: []ConditionField{
		{Op: EqOp, ConditionValue: ConditionValue{Name: "userGroup", Type: StringType, Value: `tes"t'er`}},
	}}

	cond2Exp = Condition{Op: OrOp, Conditions: []Condition{
		{Fields: []ConditionField{
			{ConditionValue: ConditionValue{Name: "useriD", Type: IntType, Value: int64(10)}, Op: LtOp},
			{ConditionValue: ConditionValue{Name: "s", Type: BoolType, Value: true}},
		}},
		{Fields: []ConditionField{
			{ConditionValue: ConditionValue{Name: "foo", Type: StringType, Value: "bar"}},
			{ConditionValue: ConditionValue{Name: "alpha", Type: IntType, Value: int64(14)}},
		}},
		{Fields: []ConditionField{
			{ConditionValue: ConditionValue{Name: "test", Type: FloatType, Value: 42.1}},
			{ConditionValue: ConditionValue{Name: "test2", Type: BoolType, Value: false}},
		}},
	}}
)

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
		{name: "stringLit 3", in: `"some \'string"`, want: []*token{{kind: stringLit, pos: 0, val: []byte("some \\'string"), opened: '"'}}},
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
		{name: "invalid float", in: "14.1.2", wantErr: true},
		{name: "invalid char", in: "@", wantErr: true},
		{name: ">", in: ">", want: []*token{{kind: gtOp}}},
		{name: "complex string", in: `'some > string \< with \' \" data |= &! @% \\ ()'`, want: []*token{{kind: stringLit, val: []byte(`some > string \< with ' \" data |= &! @% \\ ()`), opened: '\''}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lexer(strings.NewReader(tt.in))
			if (err != nil) != tt.wantErr {
				t.Errorf("lexer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestParseCondition(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    Condition
		wantErr bool
	}{
		{name: "invalid field", in: "foo true", wantErr: true},
		{name: "invalid parent 1", in: "foo != true)", wantErr: true},
		{name: "invalid parent 2", in: "(foo != true", wantErr: true},
		{name: "invalid op", in: "foo != < true", wantErr: true},
		{name: "single field", in: "foo != true", want: Condition{Fields: []ConditionField{
			{Op: NeOp, ConditionValue: ConditionValue{Name: "foo", Type: BoolType, Value: true}},
		}}},
		{name: "two field 1", in: "foo != true && bar < 20", want: Condition{Fields: []ConditionField{
			{Op: NeOp, ConditionValue: ConditionValue{Name: "foo", Type: BoolType, Value: true}},
			{Op: LtOp, ConditionValue: ConditionValue{Name: "bar", Type: IntType, Value: int64(20)}},
		}}},
		{name: "two field 2", in: "foo != true || bar < 20", want: Condition{Fields: []ConditionField{
			{Op: NeOp, ConditionValue: ConditionValue{Name: "foo", Type: BoolType, Value: true}},
			{Op: LtOp, ConditionValue: ConditionValue{Name: "bar", Type: IntType, Value: int64(20)}},
		}, Op: OrOp}},
		{name: "cond1", in: cond1, want: cond1Exp},
		{name: "cond2", in: cond2, want: cond2Exp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCondition(strings.NewReader(tt.in))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
